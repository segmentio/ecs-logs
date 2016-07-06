package lib

import (
	"encoding/json"

	"github.com/segmentio/ecs-logs-go"
)

type Message struct {
	Group  string        `json:"group,omitempty"`
	Stream string        `json:"stream,omitempty"`
	Event  ecslogs.Event `json:"event,omitempty"`
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

type MessageBatch []Message

func (list MessageBatch) Swap(i int, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list MessageBatch) Less(i int, j int) bool {
	return list[i].Event.Time.Before(list[j].Event.Time)
}

func (list MessageBatch) Len() int {
	return len(list)
}
