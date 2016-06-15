package ecslogs

import (
	"reflect"
	"testing"
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
