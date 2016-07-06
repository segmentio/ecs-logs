package ecslogs

import (
	"bytes"
	"testing"
)

func TestLoggerLog(t *testing.T) {
	tests := []struct {
		e Event
		s string
	}{
		{
			e: Eprintf(DEBUG, "Hello %s!", "World"),
			s: `{"level":"DEBUG","time":"0001-01-01T00:00:00Z","info":{},"data":{},"message":"Hello World!"}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLogger(b)
		log.Log(test.e)

		if s := b.String(); s != test.s {
			t.Errorf("\n- expected: %s\n- found:    %s", test.s, s)
		}
	}
}
