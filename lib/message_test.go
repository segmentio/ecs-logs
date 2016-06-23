package ecslogs

import (
	"errors"
	"io"
	"reflect"
	"testing"
	"time"
)

func TestMessageString(t *testing.T) {
	m := Message{
		Level:   INFO,
		Group:   "abc",
		Stream:  "0123456789",
		Content: "Hello World!",
		Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
	}

	const ref = `{"level":"INFO","group":"abc","stream":"0123456789","content":"Hello World!","time":"2016-06-13T12:23:42.123456789Z"}`

	if s := m.String(); s != ref {
		t.Errorf("invalid string representation of the message:\n - expected: %s\n - found:    %s", ref, s)
	}
}

func TestMessageEncoderDecover(t *testing.T) {
	batch := []Message{
		Message{
			Level:   INFO,
			Group:   "abc",
			Stream:  "0123456789",
			Content: "Hello World!",
			Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
		},
		Message{
			Level:   INFO,
			Group:   "abc",
			Stream:  "0123456789",
			Content: "How are you doing?",
			Time:    time.Date(2016, 6, 13, 12, 24, 42, 123456789, time.UTC),
		},
	}

	r, w := io.Pipe()
	e := NewMessageEncoder(w)
	d := NewMessageDecoder(r)

	// This goroutine encodes the batch of message using the encoder that
	// outputs the serialized messages to the write end of the pipe.
	go func() {
		if err := e.WriteMessageBatch(batch); err != nil {
			t.Error(err)
			return
		}

		if err := e.Close(); err != nil {
			t.Error(err)
			return
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
			Level:   INFO,
			Group:   "abc",
			Stream:  "0123456789",
			Content: "Hello World!",
			Time:    time.Date(2016, 6, 13, 12, 23, 42, 123456789, time.UTC),
		},
		Message{
			Level:   INFO,
			Group:   "abc",
			Stream:  "0123456789",
			Content: "How are you doing?",
			Time:    time.Date(2016, 6, 13, 12, 24, 42, 123456789, time.UTC),
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
