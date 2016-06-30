package ecslogs

import (
	"bytes"
	"testing"
)

func TestLoggerPrintf(t *testing.T) {
	tests := []struct {
		method  func(*Logger, string, ...interface{})
		format  string
		args    []interface{}
		message string
	}{
		{
			method: (*Logger).Debugf,
			format: "Hello %s!",
			args:   []interface{}{"World"},
			message: `{"level":"DEBUG","message":"Hello World!","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLoggerWith(LoggerConfig{
			Output: NewLoggerOutput(b),
			Caller: testCaller,
		})
		test.method(log, test.format, test.args...)

		if s := b.String(); s != test.message {
			t.Errorf("\n- expected: %#v\n- found:    %#v", test.message, s)
		}
	}
}

func TestLoggerPrint(t *testing.T) {
	tests := []struct {
		method  func(*Logger, ...interface{})
		args    []interface{}
		message string
	}{
		{
			method: (*Logger).Debug,
			message: `{"level":"DEBUG","message":"","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLoggerWith(LoggerConfig{
			Output: NewLoggerOutput(b),
			Caller: testCaller,
		})
		test.method(log, test.args...)

		if s := b.String(); s != test.message {
			t.Errorf("\n- expected: %#v\n- found:    %#v", test.message, s)
		}
	}
}

func TestLoggerWith(t *testing.T) {
	tests := []struct {
		event   Event
		message string
	}{
		{
			event: nil,
			message: `{"level":"DEBUG","message":"the log message","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"}
`,
		},
		{
			event: NewEvent(),
			message: `{"level":"DEBUG","message":"the log message","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"}
`,
		},
		{
			event: NewEvent(Tag{"hello", "world"}),
			message: `{"hello":"world","level":"DEBUG","message":"the log message","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLoggerWith(LoggerConfig{
			Output: NewLoggerOutput(b),
			Caller: testCaller,
		})
		log.With(test.event).Debug("the log message")

		if s := b.String(); s != test.message {
			t.Errorf("\n- expected: %#v\n- found:    %#v", test.message, s)
		}
	}
}

func testCaller(_ int) (string, int, string, bool) {
	return "github.com/segmentio/ecs-logs/lib/logger_test.go", 42, "F", true
}
