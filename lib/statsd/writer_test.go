package statsd

import (
	"strings"
	"testing"

	"github.com/segmentio/ecs-logs/lib"
)

func TestExtractMetrics(t *testing.T) {
	batch := ecslogs.MessageBatch{
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.WARN},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.ERROR},
		},
		ecslogs.Message{
			Event: ecslogs.Event{Level: ecslogs.WARN},
		},
	}

	metrics := extractMetrics(batch)

	if len(metrics) != 3 {
		t.Errorf("invalid metrics count: %d != %d", len(metrics), 3)
	}

	countInfo := 0
	countWarn := 0
	countError := 0

	for lvl, m := range metrics {
		switch lvl {
		case ecslogs.INFO:
			countInfo += m.value

		case ecslogs.WARN:
			countWarn += m.value

		case ecslogs.ERROR:
			countError += m.value

		default:
			t.Errorf("invalid metric level: %s", lvl)
			continue
		}

		if m.name != strings.ToLower(lvl.String()) {
			t.Errorf("invalid metric name for level %s: %s", lvl, m.name)
		}
	}

	if countInfo != 4 {
		t.Error("invalid info count:", countInfo)
	}

	if countWarn != 2 {
		t.Error("invalid warning count:", countWarn)
	}

	if countError != 1 {
		t.Error("invalid error count:", countError)
	}
}
