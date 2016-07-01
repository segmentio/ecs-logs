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
			s: `{"level":"INFO","time":"0001-01-01T00:00:00Z","info":{},"data":{},"message":"answer = 42"}`,
		},
		{
			e: Eprintf(WARN, "an error was raised (%s)", syscall.Errno(2)),
			s: `{"level":"WARN","time":"0001-01-01T00:00:00Z","info":{"errors":[{"type":"syscall.Errno","error":"no such file or directory","errno":2}]},"data":{},"message":"an error was raised (no such file or directory)"}`,
		},
		{
			e: Eprint(ERROR, "an error was raised:", io.EOF),
			s: `{"level":"ERROR","time":"0001-01-01T00:00:00Z","info":{"errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{},"message":"an error was raised: EOF"}`,
		},
	}

	for _, test := range tests {
		if s := test.e.String(); s != test.s {
			t.Errorf("\n- expected: %s\n- found:    %s", test.s, s)
		}
	}
}
