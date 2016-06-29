// +build linux

package journald

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/segmentio/ecs-logs/lib"
)

func NewReader() (r ecslogs.Reader, err error) {
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

func (r reader) ReadMessage() (msg ecslogs.Message, err error) {
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

func (r reader) getMessage() (msg ecslogs.Message, ok bool, err error) {
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

	msg.Level = r.getPriority()
	msg.PID = r.getInt("_PID")
	msg.UID = r.getInt("_UID")
	msg.GID = r.getInt("_GID")
	msg.Errno = r.getInt("ERRNO")
	msg.Line = r.getInt("CODE_LINE")
	msg.Func = r.getString("CODE_FUNC")
	msg.File = r.getString("CODE_FILE")
	msg.ID = r.getString("MESSAGE_ID")
	msg.Host = r.getString("_HOSTNAME")
	msg.Content = r.getContent("MESSAGE")
	msg.Time = r.getTime()
	ok = true
	return
}

func (r reader) getInt(k string) (v int) {
	v, _ = strconv.Atoi(r.getString(k))
	return
}

func (r reader) getTime() (t ecslogs.Timestamp) {
	if u, e := r.GetRealtimeUsec(); e == nil {
		t = ecslogs.MakeTimestamp(time.Unix(int64(u/1000000), int64((u%1000000)*1000)))
	}
	return
}

func (r reader) getPriority() (p ecslogs.Level) {
	if v, e := strconv.Atoi(r.getString("PRIORITY")); e != nil {
		p = ecslogs.INFO
	} else {
		p = ecslogs.Level(v)
	}
	return
}

func (r reader) getString(k string) (s string) {
	s, _ = r.GetDataValue(k)
	return
}

func (r reader) getContent(k string) (v interface{}) {
	var s string
	var e error

	if s, e = r.GetDataValue(); e != nil {
		return
	}

	if json.Unmarshal([]byte(s), &v) != nil {
		v = s
	}

	return
}
