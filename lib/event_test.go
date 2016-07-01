package ecslogs

import (
	"io"
	"syscall"
	"testing"
)

func TestEvent(t *testing.T) {
	tests := []struct {
		e Event
		s string
	}{
		{
			e: Eprintf(INFO, "answer = %d", 42),
			s: `{"info":{"level":"INFO"},"data":{"message":"answer = 42"}}`,
		},
		{
			e: Eprintf(WARN, "an error was raised (%s)", syscall.Errno(2)),
			s: `{"info":{"level":"WARN","errors":[{"type":"syscall.Errno","error":"no such file or directory","errno":2}]},"data":{"message":"an error was raised (no such file or directory)"}}`,
		},
		{
			e: Eprint(ERROR, "an error was raised:", io.EOF),
			s: `{"info":{"level":"ERROR","errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{"message":"an error was raised: EOF"}}`,
		},
	}

	for _, test := range tests {
		if s := test.e.String(); s != test.s {
			t.Errorf("\n- expected: %s\n- found:    %s", s, test.s)
		}
	}
}
