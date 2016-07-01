package ecslogs

import "testing"

func TestParseFunctionName(t *testing.T) {
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
	}

	for _, test := range tests {
		if pkg, fn := parseFunctionName(test.name); pkg != test.pkg || fn != test.fn {
			t.Errorf("%s => %s - %s", test.name, pkg, fn)
		}
	}
}
