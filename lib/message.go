package ecslogs

import (
	"encoding/json"
	"strconv"
	"strings"
)

type Message struct {
	Group  string `json:"group,omitempty"`
	Stream string `json:"stream,omitempty"`
	Event  Event  `json:"event,omitempty"`
}

func (m Message) Bytes() []byte {
	b, _ := json.Marshal(m)
	return b
}

func (m Message) String() string {
	return string(m.Bytes())
}

func (m Message) ContentLength() int {
	return jsonLen(m.Event)
}

func MessageSource(file string, line int, function string) string {
	parts := make([]string, 0, 3)

	if len(file) != 0 {
		parts = append(parts, file)
	}

	if line != 0 {
		parts = append(parts, strconv.Itoa(line))
	}

	if len(function) != 0 {
		parts = append(parts, function)
	}

	return strings.Join(parts, ":")
}
