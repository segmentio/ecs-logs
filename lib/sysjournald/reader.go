// +build linux

package sysjournald

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

	var priorityEnv string
	if priorityEnv = os.Getenv("SYSTEM_JOURNALD_PRIORITY"); len(priorityEnv) == 0 {
		priorityEnv = "NOTICE"
	}

	var priority ecslogs.Level
	if priority, err = ecslogs.ParseLevel(priorityEnv); err != nil {
		return
	}

	var group string
	if group = os.Getenv("SERVER_GROUP"); len(group) == 0 {
		err = fmt.Errorf("missing SERVER_GROUP environment variable")
		return
	}

	r = &reader{Journal: j, priority: priority, group: group}
	return
}

type reader struct {
	priority ecslogs.Level
	group    string
	stopped  int32
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
	if tag, _ := r.GetDataValue("CONTAINER_TAG"); len(tag) != 0 {
		// Found CONTAINER_TAG, logs from docker containers are not handle here.
		return
	}

	if msg.Event.Level == ecslogs.NONE {
		msg.Event.Level = r.getPriority()
	}

	if msg.Event.Level >= r.priority {
		// Skip messages which don't have the correct priority
		err = nil
		return
	}

	msg.Group = r.group
	msg.Stream = r.getString("_HOSTNAME")

	if s := r.getString("MESSAGE"); len(s) != 0 {
		d := json.NewDecoder(strings.NewReader(s))
		d.UseNumber()

		if d.Decode(&msg.Event) != nil {
			msg.Event.Message = s
		}
	}

	if len(msg.Event.Info.Host) == 0 {
		msg.Event.Info.Host = r.getString("_HOSTNAME")
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

	msg.Event.Data = make(map[string]interface{}, 3)
	msg.Event.Data["CMDLINE"] = r.getString("_CMDLINE")
	msg.Event.Data["EXE"] = r.getString("_EXE")
	msg.Event.Data["SYSLOG_IDENFIFIER"] = r.getString("SYSLOG_IDENTIFIER")

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
