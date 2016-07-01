package ecslogs_logrus

import (
	"bytes"
	"io"
	"testing"

	"github.com/Sirupsen/logrus"
)

func TestFormatter(t *testing.T) {
	buf := &bytes.Buffer{}
	log := &logrus.Logger{
		Out:       buf,
		Formatter: NewFormatter(),
		Level:     logrus.DebugLevel,
	}

	log.
		WithError(io.EOF).
		WithField("hello", "world").
		Errorf("an error was raised (%s)", io.EOF)

	if s := buf.String(); len(s) == 0 {
		t.Error("logrus formatter failed: empty buffer")
	} else {
		// I wish we could make better testing here but the logrus
		// API doesn't let us mock the timestamp so we can't really
		// predict what "time" is gonna be.
		t.Log(s)
	}
}
