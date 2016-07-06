package statsd

import (
	"testing"

	"github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"
)

func TestExtractMetrics(t *testing.T) {
	batch := lib.MessageBatch{
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.WARN},
		},
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.INFO},
		},
		lib.Message{
			Event: ecslogs.Event{Level: ecslogs.ERROR},
		},
		lib.Message{
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
