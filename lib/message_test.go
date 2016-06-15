package ecslogs

import (
	"testing"
	"time"
)

func TestMessageString(t *testing.T) {
	m := Message{
		Group:   "abc",
		Stream:  "0123456789",
		Content: "Hello World!",
		Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
	}

	const ref = `{"group":"abc","stream":"0123456789","content":"Hello World!","time":"2016-06-13T12:23:42.123456789Z"}`

	if s := m.String(); s != ref {
		t.Errorf("invalid string representation of the message:\n - expected: %s\n - found:    %s", ref, s)
	}
}
