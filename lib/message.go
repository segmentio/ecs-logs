package lib

import (
	"encoding/json"
	"sync"

	"github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/jutil"
)

type Message struct {
	Group  string                 `json:"group,omitempty"`
	Stream string                 `json:"stream,omitempty"`
	Event  ecslogs.Event          `json:"event,omitempty"`
	JSON   map[string]interface{} `json:"json,omitempty"`
}

func (m Message) Bytes() []byte {
	b, _ := json.Marshal(m)
	return b
}

func (m Message) GoodEggsEventString() string {
	if m.JSON != nil {
		b, _ := json.Marshal(m.JSON)
		return string(b)
	}
	return m.Event.String()
}

func (m Message) String() string {
	return string(m.Bytes())
}

func (m Message) ContentLength() int {
	n, _ := jutil.Length(m.Event)
	return n
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

type MessageQueue struct {
	C      <-chan struct{}
	signal chan struct{}
	mutex  sync.Mutex
	batch  MessageBatch
}

func NewMessageQueue() *MessageQueue {
	c := make(chan struct{}, 1)
	return &MessageQueue{
		C:      c,
		signal: c,
		batch:  make(MessageBatch, 0, 100),
	}
}

func (q *MessageQueue) Push(msg Message) {
	q.mutex.Lock()
	q.batch = append(q.batch, msg)
	q.mutex.Unlock()
}

func (q *MessageQueue) Notify() {
	select {
	default:
	case q.signal <- struct{}{}:
	}
}

func (q *MessageQueue) Flush() (batch MessageBatch) {
	q.mutex.Lock()
	batch = make(MessageBatch, len(q.batch))
	copy(batch, q.batch)
	q.batch = q.batch[:0]
	q.mutex.Unlock()
	return
}
