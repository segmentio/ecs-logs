package ecslogs

import (
	"encoding/json"
	"fmt"
)

type Message struct {
	Level   Level     `json:"level,omitempty"`
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
	Content Content   `json:"content"`
	Time    Timestamp `json:"time,omitempty"`
}

func (m Message) String() string {
	return fmt.Sprint(m)
}

func (m Message) Format(f fmt.State, _ rune) {
	b, _ := json.Marshal(m)
	f.Write(b)
}

func (m *Message) ExtractContentMetadata() {
	var tmp Message

	if m.Content.Value == nil {
		return
	}

	if json.Unmarshal(m.Content.Raw, &tmp) != nil {
		return
	}

	if tmp.Level != 0 {
		m.Level = tmp.Level
		delete(tmp.Content.Value, "level")
	}

	if tmp.PID != 0 {
		m.PID = tmp.PID
		delete(tmp.Content.Value, "pid")
	}

	if tmp.UID != 0 {
		m.UID = tmp.UID
		delete(tmp.Content.Value, "uid")
	}

	if tmp.GID != 0 {
		m.GID = tmp.GID
		delete(tmp.Content.Value, "gid")
	}

	if tmp.Errno != 0 {
		m.Errno = tmp.Errno
		delete(tmp.Content.Value, "errno")
	}

	if tmp.Line != 0 {
		m.Line = tmp.Line
		delete(tmp.Content.Value, "line")
	}

	if len(tmp.Func) != 0 {
		m.Func = tmp.Func
		delete(tmp.Content.Value, "func")
	}

	if len(tmp.File) != 0 {
		m.File = tmp.File
		delete(tmp.Content.Value, "file")
	}

	if len(tmp.ID) != 0 {
		m.ID = tmp.ID
		delete(tmp.Content.Value, "id")
	}

	// Host, Group, Stream, Content and Time cannot be overwritten by the fields
	// of the content object.
	return
}
