package ecslogs_logrus

import (
	"bytes"
	"io"
	"strings"
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

	s := buf.String()

	// I wish we could make better testing here but the logrus
	// API doesn't let us mock the timestamp so we can't really
	// predict what "time" is gonna be.
	if !strings.HasPrefix(s, `{"level":"ERROR","time":"`) || !strings.HasSuffix(s, `","info":{"errors":[{"type":"*errors.errorString","error":"EOF"}]},"data":{"hello":"world"},"message":"an error was raised (EOF)"}
`) {
		t.Error("logrus formatter failed:", s)
	}
}
