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
		t.Logf("testing split of %v at %d", test.list, test.count)
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
	st := NewStream("0123456789", ts)

	if s := st.Name(); s != "0123456789" {
		t.Error("invalid stream name:", s)
	}
}

func TestStreamExpired(t *testing.T) {
	ts := time.Now()
	st := NewStream("0123456789", ts)

	if !st.HasExpired(1*time.Second, ts.Add(2*time.Second)) {
		t.Error("new stream should be considered expired because it has no messages and wasn't updated recently")
	}
}

func TestStreamNotExpiredDueToMessages(t *testing.T) {
	ts := time.Now()
	st := NewStream("0123456789", ts)
	st.Add(Message{
		Group:   "abc",
		Stream:  "0123456789",
		Content: "Hello World!",
		Time:    ts,
	}, ts)

	if st.HasExpired(1*time.Second, ts.Add(2*time.Second)) {
		t.Error("non-empty streams should not be expired")
	}
}

func TestStreamNotExpiredDueToTimeout(t *testing.T) {
	ts := time.Now()
	st := NewStream("0123456789", ts)

	if st.HasExpired(1*time.Second, ts) {
		t.Error("streams that were updated recently should not be expired")
	}
}
