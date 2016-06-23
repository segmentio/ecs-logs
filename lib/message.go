package ecslogs

import (
	"encoding/json"
	"fmt"
	"time"
)

type Message struct {
	Level   Level     `json:"level"`
	PID     int       `json:"pid,omitempty"`
	UID     int       `json:"uid,omitempty"`
	GID     int       `json:"gid,omitempty"`
	Errno   int       `json:"errno,omitempty"`
	Line    int       `json:"line,omitempty"`
	Func    string    `json:"func,omitempty"`
	File    string    `json:"file,omitempty"`
	ID      string    `json:"id,omitempty"`
	Host    string    `json:"host,omitempty"`
	Group   string    `json:"group,omitempty"`
	Stream  string    `json:"stream,omitempty"`
	Content string    `json:"content,omitempty"`
	Time    time.Time `json:"time,omitempty"`
}

func (m Message) String() string {
	return fmt.Sprint(m)
}

func (m Message) Format(f fmt.State, _ rune) {
	b, _ := json.Marshal(m)
	f.Write(b)
}
