package lib

import (
	"encoding/json"
	"io"
)

type Writer interface {
	io.Closer

	WriteMessage(Message) error

	WriteMessageBatch(MessageBatch) error
}

func NewMessageEncoder(w io.Writer) Writer {
	return encoder{
		j: json.NewEncoder(w),
		w: w,
	}
}

type encoder struct {
	j *json.Encoder
	w io.Writer
}

func (e encoder) Close() (err error) {
	return
}

func (e encoder) WriteMessage(msg Message) (err error) {
	err = e.j.Encode(msg)
	return
}

func (e encoder) WriteMessageBatch(batch MessageBatch) (err error) {
	for _, msg := range batch {
		if err = e.WriteMessage(msg); err != nil {
			return
		}
	}
	return
}
