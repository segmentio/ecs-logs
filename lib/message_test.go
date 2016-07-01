package ecslogs

import (
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestMessageSource(t *testing.T) {
	tests := []struct {
		file   string
		line   int
		fn     string
		source string
	}{
		{
			file:   "test.go",
			line:   42,
			fn:     "F",
			source: "test.go:42:F",
		},
	}

	for _, test := range tests {
		if s := MessageSource(test.file, test.line, test.fn); s != test.source {
			t.Errorf("invalid source: %#v", s)
		}
	}
}

func TestMessageString(t *testing.T) {
	d := time.Date(2016, 6, 13, 12, 23, 42, 123456000, time.Local)
	m := Message{
		Group:  "abc",
		Stream: "0123456789",
		Event: Event{
			Level:   INFO,
			Time:    d,
			Message: "Hello World!",
			Info:    EventInfo{Host: "localhost"},
			Data:    EventData{},
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

func TestMessageEncoderDecoder(t *testing.T) {
	batch := []Message{
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: Event{
				Level:   INFO,
				Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
				Message: "Hello World!",
			},
		},
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: Event{
				Level:   INFO,
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
	var list []Message
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
	batch := []Message{
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: Event{
				Level:   INFO,
				Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
				Message: "Hello World!",
			},
		},
		Message{
			Group:  "abc",
			Stream: "0123456789",
			Event: Event{
				Level:   INFO,
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
