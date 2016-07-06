package ecslogs

import "testing"

func TestParseFuncName(t *testing.T) {
	tests := []struct {
		name string
		pkg  string
		fn   string
	}{
		{
			name: "github.com/segmentio/ecs-logs/lib.TestLoggerPrintf",
			pkg:  "github.com/segmentio/ecs-logs/lib",
			fn:   "TestLoggerPrintf",
		},
		{
			name: "github.com/segmentio/ecs-logs/lib.(*Logger).Printf",
			pkg:  "github.com/segmentio/ecs-logs/lib",
			fn:   "(*Logger).Printf",
		},
		{
			name: "github.com/segmentio/ecs-logs/lib/apex.NewHandlerWith.func1",
			pkg:  "github.com/segmentio/ecs-logs/lib/apex",
			fn:   "NewHandlerWith.func1",
		},
		{
			name: "github.com/segmentio/ecs-logs/vendor/github.com/apex/log.HandlerFunc.HandleLog",
			pkg:  "github.com/segmentio/ecs-logs/vendor/github.com/apex/log",
			fn:   "HandlerFunc.HandleLog",
		},
	}

	for _, test := range tests {
		if pkg, fn := parseFuncName(test.name); pkg != test.pkg || fn != test.fn {
			t.Errorf("%s => %s - %s", test.name, pkg, fn)
		}
	}
}

func TestFuncInfoString(t *testing.T) {
	tests := []struct {
		info   FuncInfo
		source string
	}{
		{
			info:   FuncInfo{File: "test.go", Func: "F", Line: 42},
			source: "test.go:42:F",
		},
	}

	for _, test := range tests {
		if s := test.info.String(); s != test.source {
			t.Errorf("invalid source: %#v", s)
		}
	}
}

func BenchmarkGuessCaller(b *testing.B) {
	for i := 0; i != b.N; i++ {
		GuessCaller(1, 10, "github.com/segmentio/ecs-logs")
	}
}
