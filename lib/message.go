package ecslogs

import "encoding/json"

type Message struct {
	Level   Level       `json:"level,omitempty"`
	PID     int         `json:"pid,omitempty"`
	UID     int         `json:"uid,omitempty"`
	GID     int         `json:"gid,omitempty"`
	Errno   int         `json:"errno,omitempty"`
	Line    int         `json:"line,omitempty"`
	Func    string      `json:"func,omitempty"`
	File    string      `json:"file,omitempty"`
	ID      string      `json:"id,omitempty"`
	Host    string      `json:"host,omitempty"`
	Group   string      `json:"group,omitempty"`
	Stream  string      `json:"stream,omitempty"`
	Content interface{} `json:"content,omitempty"`
	Time    Timestamp   `json:"time,omitempty"`
}

func ParseMessage(s string) (m Message, err error) {
	if err = json.Unmarshal([]byte(s), &m); err != nil {
		return
	}
	m.ExtractContentMetadata()
	return
}

func (m *Message) ExtractContentMetadata() {
	switch c := m.Content.(type) {
	case map[string]interface{}:
		for k, v := range c {
			switch k {
			case "level":
				if lvl, ok := levelValue(v); ok {
					m.Level = lvl
					delete(c, k)
				}

			case "pid":
				if pid, ok := intValue(v); ok {
					m.PID = pid
					delete(c, k)
				}

			case "uid":
				if uid, ok := intValue(v); ok {
					m.UID = uid
					delete(c, k)
				}

			case "gid":
				if uid, ok := intValue(v); ok {
					m.UID = uid
					delete(c, k)
				}

			case "errno":
				if errno, ok := intValue(v); ok {
					m.Errno = errno
					delete(c, k)
				}

			case "line":
				if line, ok := intValue(v); ok {
					m.Line = line
					delete(c, k)
				}

			case "func":
				if fn, ok := stringValue(v); ok {
					m.Func = fn
					delete(c, k)
				}

			case "file":
				if file, ok := stringValue(v); ok {
					m.File = file
					delete(c, k)
				}

			case "id":
				if id, ok := stringValue(v); ok {
					m.ID = id
					delete(c, k)
				}

			default:
				// Host, Group, Stream, Content and Time cannot be overwritten
				// by the fields the content object.
			}
		}
	}
}

func (m Message) String() string {
	b, _ := json.Marshal(m)
	return string(b)
}

func (m Message) ContentLength() int {
	// TOOD: optimize this so we don't actually serialize the content and just
	// compute the length.
	b, _ := json.Marshal(m.Content)
	return len(b)
}
