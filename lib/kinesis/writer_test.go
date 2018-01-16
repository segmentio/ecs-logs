package kinesis

import (
	"testing"

	ecslogs "github.com/segmentio/ecs-logs-go"
	"github.com/segmentio/ecs-logs/lib"
)

func TestWriter(t *testing.T) {
	group, stream := "myGroup", "myStream"
	w, err := NewWriter(group, stream)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	m := lib.Message{
		Group:  group,
		Stream: stream,
		Event:  ecslogs.MakeEvent(ecslogs.DEBUG, "test"),
	}
	if err := w.WriteMessage(m); err != nil {
		t.Error(err)
	}

	if err := w.WriteMessageBatch(lib.MessageBatch{m, m, m, m, m}); err != nil {
		t.Error(err)
	}
}
