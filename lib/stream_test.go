package ecslogs

import (
	"reflect"
	"testing"
	"time"
)

func TestSplitMessageListHead(t *testing.T) {
	tests := []struct {
		list  []Message
		count int
	}{
		{
			list:  []Message{},
			count: 0,
		},
		{
			list: []Message{
				Message{Group: "A"},
				Message{Group: "B"},
				Message{Group: "C"},
			},
			count: 0,
		},
		{
			list: []Message{
				Message{Group: "A"},
				Message{Group: "B"},
				Message{Group: "C"},
			},
			count: 1,
		},
		{
			list: []Message{
				Message{Group: "A"},
				Message{Group: "B"},
				Message{Group: "C"},
			},
			count: 3,
		},
	}

	for _, test := range tests {
		head, tail := splitMessageListHead(test.list, test.count)

		if !reflect.DeepEqual(head, test.list[:test.count]) {
			t.Errorf("invalid head:\n- expected: %v\n- found:    %v", test.list[:test.count], head)
		}

		if !reflect.DeepEqual(tail, test.list[test.count:]) {
			t.Errorf("invalid tail:\n- expected: %v\n- found:    %v", test.list[test.count:], tail)
		}
	}
}

func TestStreamName(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)

	if s := st.Name(); s != "0123456789" {
		t.Error("invalid stream name:", s)
	}
}

func TestStreamString(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)

	if s := st.String(); s != `stream { group = "A", name = "0123456789" }` {
		t.Error("invalid stream name:", s)
	}
}

func TestStreamExpired(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)

	if !st.HasExpired(1*time.Second, ts.Add(2*time.Second)) {
		t.Error("new stream should be considered expired because it has no messages and wasn't updated recently")
	}
}

func TestStreamNotExpiredDueToMessages(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)
	st.Add(Message{
		Group:  "A",
		Stream: "0123456789",
		Event: Event{
			Info: EventInfo{Time: MakeTimestamp(ts)},
			Data: EventData{"message": "Hello World!"},
		},
	}, ts)

	if st.HasExpired(1*time.Second, ts.Add(2*time.Second)) {
		t.Error("non-empty streams should not be expired")
	}
}

func TestStreamNotExpiredDueToTimeout(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)

	if st.HasExpired(1*time.Second, ts) {
		t.Error("streams that were updated recently should not be expired")
	}
}

func TestStreamBytes(t *testing.T) {
	ts := time.Now()
	st := NewStream("A", "0123456789", ts)
	m1 := Message{
		Group:  "A",
		Stream: "0123456789",
		Event: Event{
			Info: EventInfo{Time: MakeTimestamp(ts)},
			Data: EventData{"message": "Hello World!"},
		},
	}
	m2 := Message{
		Group:  "A",
		Stream: "0123456789",
		Event: Event{
			Info: EventInfo{Time: MakeTimestamp(ts)},
			Data: EventData{"message": "How are you?"},
		},
	}
	m3 := Message{
		Group:  "A",
		Stream: "0123456789",
		Event: Event{
			Info: EventInfo{Time: MakeTimestamp(ts)},
			Data: EventData{"message": "Well"},
		},
	}

	st.Add(m1, ts)
	st.Add(m2, ts)
	st.Add(m3, ts)

	if bytes := m1.ContentLength() + m2.ContentLength() + m3.ContentLength(); bytes != st.bytes {
		t.Errorf("invalid stream bytes count: %d != %d", st.bytes, bytes)
	}

	// Flush the first two messages, there should be one message left in the
	// stream and the number of bytes should match the length of the third
	// message.
	list, _ := st.Flush(StreamLimits{
		MaxBytes: m1.ContentLength() + m2.ContentLength(),
	}, ts)

	if !reflect.DeepEqual(list, []Message{m1, m2}) {
		t.Error("invalid list of messages flushed from stream:", list)
	}

	if !reflect.DeepEqual(st.messages, []Message{m3}) {
		t.Error("invalid list of messages left in stream:", st.messages)
	}

	if st.bytes != m3.ContentLength() {
		t.Error("invalid stream bytes count left in stream:", st.bytes)
	}

	// Flush the last message with a limit that's shorter than the current byte
	// count in the stream.
	list, _ = st.Flush(StreamLimits{
		MaxBytes: m3.ContentLength() - 1,
	}, ts)

	if !reflect.DeepEqual(list, []Message{m3}) {
		t.Error("invalid list of messages flushed from stream:", list)
	}

	if len(st.messages) != 0 {
		t.Error("invalid list of messages left in stream:", st.messages)
	}

	if st.bytes != 0 {
		t.Error("invalid stream bytes count left in stream:", st.bytes)
	}
}
