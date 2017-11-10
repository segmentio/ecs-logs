package lib

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/segmentio/ecs-logs-go"
)

func TestMessageString(t *testing.T) {
	d := time.Date(2016, 6, 13, 12, 23, 42, 123456000, time.Local)
	m := Message{
		Group:  "abc",
		Stream: "0123456789",
		Event: ecslogs.Event{
			Level:   ecslogs.INFO,
			Time:    d,
			Message: "Hello World!",
			Info:    ecslogs.EventInfo{Host: "localhost"},
			Data:    ecslogs.EventData{},
		},
	}

	ref := fmt.Sprintf(
		`{"group":"abc","stream":"0123456789","event":{"level":"INFO","time":"%s","info":{"host":"localhost"},"data":{},"message":"Hello World!"}}`,
		d.Format(time.RFC3339Nano),
	)

	if s := m.String(); s != ref {
		t.Errorf("invalid string representation of the message:\n - expected: %s\n - found:    %s", ref, s)
	}
}

func TestGoodEggsEventStringWithJSON(t *testing.T) {
	d := time.Date(2016, 6, 13, 12, 23, 42, 123456000, time.Local)
	json := make(map[string]interface{})
	json["msg"] = "hello world"
	m := Message{
		Group:  "abc",
		Stream: "0123456789",
		Event: ecslogs.Event{
			Level:   ecslogs.INFO,
			Time:    d,
			Message: "Hello World!",
			Info:    ecslogs.EventInfo{Host: "localhost"},
			Data:    ecslogs.EventData{},
		},
		JSON: json,
	}

	ref := `{"msg":"hello world"}`

	if s := m.GoodEggsEventString(); s != ref {
		t.Errorf("invalid string representation of the message:\n - expected: %s\n - found:    %s", ref, s)
	}
}

func TestGoodEggsEventStringWithoutJSON(t *testing.T) {
	d := time.Date(2016, 6, 13, 12, 23, 42, 123456000, time.Local)
	m := Message{
		Group:  "abc",
		Stream: "0123456789",
		Event: ecslogs.Event{
			Level:   ecslogs.INFO,
			Time:    d,
			Message: "Hello World!",
			Info:    ecslogs.EventInfo{Host: "localhost"},
			Data:    ecslogs.EventData{},
		},
	}

	ref := fmt.Sprintf(
		`{"level":"INFO","time":"%s","info":{"host":"localhost"},"data":{},"message":"Hello World!"}`,
		d.Format(time.RFC3339Nano),
	)

	if s := m.GoodEggsEventString(); s != ref {
		t.Errorf("invalid string representation of the message:\n - expected: %s\n - found:    %s", ref, s)
	}
}

func TestMessageEncoderDecoder(t *testing.T) {
	batch := MessageBatch{
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: ecslogs.Event{
				Level:   ecslogs.INFO,
				Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
				Message: "Hello World!",
			},
		},
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: ecslogs.Event{
				Level:   ecslogs.INFO,
				Time:    time.Date(2016, 6, 13, 12, 24, 42, 123456789, time.UTC),
				Message: "How are you doing?",
			},
		},
	}

	r, w := io.Pipe()
	e := NewMessageEncoder(w)
	d := NewMessageDecoder(r)

	// This goroutine encodes the batch of message using the encoder that
	// outputs the serialized messages to the write end of the pipe.
	go func() {
		defer w.Close()
		if err := e.WriteMessageBatch(batch); err != nil {
			t.Error(err)
		}
	}()

	// This loop reads the messages written to the pipe and rebuilds a list
	// of messages until EOF is reached. The orignal batch and list are then
	// compred to ensure they are the same.
	var list MessageBatch
	for {
		if msg, err := d.ReadMessage(); err != nil {
			if err == io.EOF {
				break
			}
			t.Error(err)
			return
		} else {
			list = append(list, msg)
		}
	}

	if err := d.Close(); err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(batch, list) {
		t.Errorf("invalid list decoded after encoding batch:\n- %v\n- %v", batch, list)
	}
}

func TestMessageEncoderWriteMessageBatchError(t *testing.T) {
	batch := MessageBatch{
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: ecslogs.Event{
				Level:   ecslogs.INFO,
				Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
				Message: "Hello World!",
			},
		},
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: ecslogs.Event{
				Level:   ecslogs.INFO,
				Time:    time.Date(2016, 6, 13, 12, 24, 42, 123456789, time.UTC),
				Message: "How are you doing?",
			},
		},
	}

	x := errors.New("ERR")
	e := NewMessageEncoder(errorWriter{x})

	if err := e.WriteMessageBatch(batch); err != x {
		t.Errorf("expected error (%s) but got %s", x, err)
	}
}

// The errorWriter type is used to mock message encoders with a writer that
// always returns an error so we can test error cases.
type errorWriter struct {
	err error
}

func (w errorWriter) Write(b []byte) (int, error) {
	return 0, w.err
}
