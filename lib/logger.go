package ecslogs

import (
	"encoding/json"
	"io"
	"os"
)

type Logger interface {
	Log(Event) error
}

type LoggerFunc func(Event) error

func (f LoggerFunc) Log(e Event) error {
	return f(e)
}

func NewLogger(w io.Writer) Logger {
	if w == nil {
		w = os.Stdout
	}
	enc := json.NewEncoder(w)
	return LoggerFunc(func(event Event) error { return enc.Encode(event) })
}
