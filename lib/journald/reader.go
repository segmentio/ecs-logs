// +build linux

package journald

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"
)

func NewReader() (r lib.Reader, err error) {
	var j *sdjournal.Journal

	if j, err = sdjournal.NewJournal(); err != nil {
		return
	}

	if err = j.SeekTail(); err != nil {
		j.Close()
		return
	}

	var streamName string
	if streamName = os.Getenv("JOURNALD_STREAM_NAME"); len(streamName) == 0 {
		streamName = "CONTAINER_NAME"
	}

	r = &reader{Journal: j, streamName: streamName}
	return
}

type reader struct {
	streamName string
	stopped    int32
	*sdjournal.Journal
}

func (r *reader) Close() (err error) {
	atomic.StoreInt32(&r.stopped, 1)
	return
}

func (r *reader) ReadMessage() (msg lib.Message, err error) {
	for atomic.LoadInt32(&r.stopped) == 0 {
		var cur int
		var ok bool

		if cur, err = r.Next(); err != nil {
			return
		}

		if cur == 0 {
			r.Wait(1 * time.Second)
			continue
		}

		if msg, ok, err = r.getMessage(); ok || err != nil {
			return
		}
	}

	r.Journal.Close()
	err = io.EOF
	return
}

func (r *reader) getMessage() (msg lib.Message, ok bool, err error) {
	if msg.Group, err = r.GetDataValue("CONTAINER_TAG"); len(msg.Group) == 0 {
		// No CONTAINER_TAG, this must be a journal message from a process that
		// isn't running in a docker container.
		err = nil
		return
	}

	if msg.Stream, err = r.GetDataValue(r.streamName); err != nil {
		// Fallback to CONTAINER_NAME
		if msg.Stream, err = r.GetDataValue("CONTAINER_NAME"); err != nil {

			// There's a CONTAINER_TAG but no CONTAINER_NAME, something is seriously
			// wrong here, the log docker log driver is misbehaving.
			err = fmt.Errorf("missing CONTAINER_NAME in message with CONTAINER_TAG=%s", msg.Group)

			return
		}
	}

	msg.Stream = sanitizeStreamName(msg.Stream)

	if s := r.getString("MESSAGE"); len(s) != 0 {
		// attempt to parse the line as JSON
		if json.Unmarshal([]byte(s), &msg.JSON) != nil {
			// if it isn't JSON, initialize an empty object and set it as the `msg` property
			msg.JSON = make(map[string]interface{})
			msg.JSON["msg"] = s
		}

		// ensure it has a `time` property at the root
		if msg.JSON["time"] == nil {
			msg.JSON["time"] = r.getTime()
		}

		// build up metadata object
		metadata := make(map[string]interface{})

		metadata["source"] = "stdout"
		if r.getPriority() <= 4 {
			metadata["source"] = "stderr"
		}
		metadata["host"] = r.getString("CONTAINER_ID")
		metadata["docker_id"] = r.getString("CONTAINER_ID_FULL")
		metadata["docker_name"] = r.getString("CONTAINER_NAME")
		metadata["ecs_task"] = r.getString("CONTAINER_TASK_UUID")

		// assign metadata to the `logspout` property
		msg.JSON["logspout"] = metadata
	}

	if msg.Event.Level == ecslogs.NONE {
		msg.Event.Level = r.getPriority()
	}

	if len(msg.Event.Info.Host) == 0 {
		msg.Event.Info.Host = r.getString("_HOSTNAME")
	}

	if len(msg.Event.Info.Source) == 0 {
		msg.Event.Info.Source = (ecslogs.FuncInfo{
			File: r.getString("CODE_FILE"),
			Func: r.getString("CODE_FUNC"),
			Line: r.getInt("CODE_LINE"),
		}).String()
	}

	if len(msg.Event.Info.ID) == 0 {
		msg.Event.Info.ID = r.getString("MESSAGE_ID")
	}

	if msg.Event.Info.PID == 0 {
		msg.Event.Info.PID = r.getInt("_PID")
	}

	if msg.Event.Info.GID == 0 {
		msg.Event.Info.GID = r.getInt("_GID")
	}

	if msg.Event.Info.UID == 0 {
		msg.Event.Info.UID = r.getInt("_UID")
	}

	if msg.Event.Time == (time.Time{}) {
		msg.Event.Time = r.getTime()
	}

	ok = true
	return
}

func (r *reader) getInt(k string) (v int) {
	v, _ = strconv.Atoi(r.getString(k))
	return
}

func (r *reader) getTime() (t time.Time) {
	if u, e := r.GetRealtimeUsec(); e == nil {
		t = time.Unix(int64(u/1000000), int64((u%1000000)*1000))
	} else {
		t = time.Now()
	}
	return
}

func (r *reader) getPriority() (p ecslogs.Level) {
	if v, e := strconv.Atoi(r.getString("PRIORITY")); e != nil {
		p = ecslogs.INFO
	} else {
		p = ecslogs.MakeLevel(v)
	}
	return
}

func (r *reader) getString(k string) (s string) {
	s, _ = r.GetDataValue(k)
	return
}

func sanitizeStreamName(name string) string {
	name = strings.Replace(name, ":", "/", -1)
	name = strings.Replace(name, "*", "/", -1)
	max := len(name)
	if max > 512 {
		max = 512
	}
	return name[:max]
}
