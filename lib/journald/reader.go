// +build linux

package journald

import (
	"strconv"
	"time"

	"github.com/coreos/go-systemd/sdjournal"
	"github.com/segmentio/ecs-logs/lib"
)

func NewMessageReader() (r ecslogs.MessageReadCloser, err error) {
	var j *sdjournal.Journal

	if j, err = sdjournal.NewJournal(); err != nil {
		return
	}

	if err = j.SeekTail(); err != nil {
		j.Close()
		return
	}

	r = journalReader{j}
	return
}

type journalReader struct {
	j *sdjournal.Journal
}

func (r journalReader) Close() error {
	return r.j.Close()
}

func (r journalReader) ReadMessage() (msg ecslogs.Message, err error) {
	for {
		var cur int
		var ok bool

		if cur, err = r.j.Next(); err != nil {
			return
		}

		if cur == 0 {
			r.j.Wait(sdjournal.IndefiniteWait)
			continue
		}

		if msg, ok, err = r.getMessage(); ok || err != nil {
			return
		}
	}
}

func (r journalReader) getMessage() (msg ecslogs.Message, ok bool, err error) {
	var usec uint64

	if msg.Group, _ = r.j.GetDataValue("CONTAINER_TAG"); len(msg.Group) == 0 {
		return
	}

	if msg.Stream, err = r.j.GetDataValue("CONTAINER_NAME"); err != nil {
		return
	}

	if msg.Content, err = r.j.GetDataValue("MESSAGE"); err != nil {
		return
	}

	if usec, err = r.j.GetRealtimeUsec(); err != nil {
		return
	}

	if s, e := r.j.GetDataValue("PRIORITY"); e != nil {
		msg.Level = ecslogs.INFO
	} else if v, e := strconv.Atoi(s); e != nil {
		msg.Level = ecslogs.INFO
	} else {
		msg.Level = ecslogs.Level(v)
	}

	msg.PID, _ = r.j.GetDataValue("_PID")
	msg.UID, _ = r.j.GetDataValue("_UID")
	msg.GID, _ = r.j.GetDataValue("_GID")
	msg.Errno, _ = r.j.GetDataValue("ERRNO")
	msg.Line, _ = r.j.GetDataValue("CODE_LINE")
	msg.Func, _ = r.j.GetDataValue("CODE_FUNC")
	msg.File, _ = r.j.GetDataValue("CODE_FILE")
	msg.ID, _ = r.j.GetDataValue("MESSAGE_ID")
	msg.Host, _ = r.j.GetDataValue("_HOSTNAME")
	msg.Time = time.Unix(int64(usec/1000000), int64((usec%1000000)*1000))
	ok = true
	return
}
