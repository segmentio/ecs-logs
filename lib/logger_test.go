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
			s: `{"level":"DEBUG","time":"0001-01-01T00:00:00Z","info":{"source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:TestLoggerLog"},"data":{},"message":"Hello World!"}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLogger(LoggerConfig{
			Output:   b,
			FuncInfo: testFuncInfo,
		})
		log.Log(test.e)

		if s := b.String(); s != test.s {
			t.Errorf("\n- expected: %s\n- found:    %s", test.s, s)
		}
	}
}

func testFuncInfo(pc uintptr) (info FuncInfo, ok bool) {
	if info, ok = GetFuncInfo(pc); !ok {
		return
	}
	info.Line = 42
	return
}
