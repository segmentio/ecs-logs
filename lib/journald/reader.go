// +build linux

package journald

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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

	r = reader{j}
	return
}

type reader struct {
	*sdjournal.Journal
}

func (r reader) ReadMessage() (msg lib.Message, err error) {
	for {
		var cur int
		var ok bool

		if cur, err = r.Next(); err != nil {
			return
		}

		if cur == 0 {
			r.Wait(sdjournal.IndefiniteWait)
			continue
		}

		if msg, ok, err = r.getMessage(); ok || err != nil {
			return
		}
	}
}

func (r reader) getMessage() (msg lib.Message, ok bool, err error) {
	if msg.Group, err = r.GetDataValue("CONTAINER_TAG"); len(msg.Group) == 0 {
		// No CONTAINER_TAG, this must be a journal message from a process that
		// isn't running in a docker container.
		err = nil
		return
	}

	if msg.Stream, err = r.GetDataValue("CONTAINER_NAME"); err != nil {
		// There's a CONTAINER_TAG but no CONTAINER_NAME, something is seriously
		// wrong here, the log docker log driver is misbehaving.
		err = fmt.Errorf("missing CONTAINER_NAME in message with CONTAINER_TAG=%s", msg.Group)
		return
	}

	if s := r.getString("MESSAGE"); len(s) != 0 {
		d := json.NewDecoder(strings.NewReader(s))
		d.UseNumber()

		if d.Decode(&msg.Event) != nil {
			msg.Event.Message = s
		}
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

func (r reader) getInt(k string) (v int) {
	v, _ = strconv.Atoi(r.getString(k))
	return
}

func (r reader) getTime() (t time.Time) {
	if u, e := r.GetRealtimeUsec(); e == nil {
		t = time.Unix(int64(u/1000000), int64((u%1000000)*1000))
	} else {
		t = time.Now()
	}
	return
}

func (r reader) getPriority() (p ecslogs.Level) {
	if v, e := strconv.Atoi(r.getString("PRIORITY")); e != nil {
		p = ecslogs.INFO
	} else {
		p = ecslogs.MakeLevel(v)
	}
	return
}

func (r reader) getString(k string) (s string) {
	s, _ = r.GetDataValue(k)
	return
}
