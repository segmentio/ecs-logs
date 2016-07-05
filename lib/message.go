package ecslogs

import "encoding/json"

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
