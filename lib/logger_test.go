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
			message: `{"info":{"level":"DEBUG","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"},"data":{"message":"Hello World!"}}
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
			t.Errorf("\n- expected: %s\n- found:    %s", test.message, s)
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
			message: `{"info":{"level":"DEBUG","source":"github.com/segmentio/ecs-logs/lib/logger_test.go:42:F"},"data":{"message":""}}
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
			t.Errorf("\n- expected: %s\n- found:    %s", test.message, s)
		}
	}
}

func TestLoggerWith(t *testing.T) {
	tests := []struct {
		data    interface{}
		message string
	}{
		{
			message: `{"info":{"level":"DEBUG"},"data":{"message":"the log message"}}
`,
		},

		{
			data: EventData{},
			message: `{"info":{"level":"DEBUG"},"data":{"message":"the log message"}}
`,
		},

		{
			data: EventData{"hello": "world"},
			message: `{"info":{"level":"DEBUG"},"data":{"hello":"world","message":"the log message"}}
`,
		},

		{
			data: struct{}{},
			message: `{"info":{"level":"DEBUG"},"data":{"message":"the log message"}}
`,
		},

		{
			data: struct{ Answer int }{42},
			message: `{"info":{"level":"DEBUG"},"data":{"Answer":42,"message":"the log message"}}
`,
		},

		{
			data: struct {
				Answer int `json:"answer"`
			}{42},
			message: `{"info":{"level":"DEBUG"},"data":{"answer":42,"message":"the log message"}}
`,
		},

		{
			data: struct {
				Answer int `json:",omitempty"`
			}{},
			message: `{"info":{"level":"DEBUG"},"data":{"message":"the log message"}}
`,
		},

		{
			data: struct {
				Answer int `json:"-"`
			}{},
			message: `{"info":{"level":"DEBUG"},"data":{"message":"the log message"}}
`,
		},

		{
			data: struct {
				Question string
				Answer   string
			}{"How are you?", "Well"},
			message: `{"info":{"level":"DEBUG"},"data":{"Answer":"Well","Question":"How are you?","message":"the log message"}}
`,
		},
	}

	b := &bytes.Buffer{}
	b.Grow(1024)

	for _, test := range tests {
		b.Reset()

		log := NewLoggerWith(LoggerConfig{Output: NewLoggerOutput(b)})
		log.With(test.data).Debug("the log message")

		if s := b.String(); s != test.message {
			t.Errorf("\n- expected: %s\n- found:    %s", test.message, s)
		}
	}
}

func testCaller(_ int) (string, int, string, bool) {
	return "github.com/segmentio/ecs-logs/lib/logger_test.go", 42, "F", true
}
