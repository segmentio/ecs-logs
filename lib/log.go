package lib

import (
	"github.com/apex/log"
	"github.com/segmentio/ecs-logs-go/apex"
)

type LogLevel log.Level

func (lvl *LogLevel) Set(s string) error {
	if l, e := log.ParseLevel(s); e != nil {
		return e
	} else {
		*lvl = LogLevel(l)
		return nil
	}
}

func (lvl LogLevel) Get() interface{} {
	return lvl
}

func (lvl LogLevel) String() string {
	return log.Level(lvl).String()
}

type LogHandler struct {
	Group    string
	Stream   string
	Hostname string
	Queue    *MessageQueue
}

func (h *LogHandler) HandleLog(entry *log.Entry) (err error) {
	msg := Message{
		Group:  h.Group,
		Stream: h.Stream,
		Event:  apex_ecslogs.MakeEvent(entry),
	}

	if len(msg.Event.Info.Host) == 0 {
		msg.Event.Info.Host = h.Hostname
	}

	h.Queue.Push(msg)
	h.Queue.Notify()
	return
}
