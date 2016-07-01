package logrus_ecslogs

import (
	"bytes"

	"github.com/Sirupsen/logrus"
	"github.com/segmentio/ecs-logs/lib"
)

func NewFormatter() logrus.Formatter {
	return NewFormatterWith(ecslogs.LoggerConfig{})
}

func NewFormatterWith(config ecslogs.LoggerConfig) logrus.Formatter {
	return &formatter{
		config: config,
	}
}

type formatter struct {
	config ecslogs.LoggerConfig
}

func (f *formatter) Format(entry *logrus.Entry) (b []byte, err error) {
	buf := &bytes.Buffer{}
	buf.Grow(1024)

	cfg := f.config
	cfg.Output = ecslogs.NewLoggerOutput(buf)

	log := ecslogs.NewLoggerWith(cfg)
	log.Log(makeEvent(entry))

	b = buf.Bytes()
	return
}

func makeEvent(entry *logrus.Entry) ecslogs.Event {
	return ecslogs.Event{
		Level:   makeLevel(entry.Level),
		Info:    makeEventInfo(entry),
		Data:    makeEventData(entry),
		Time:    entry.Time,
		Message: entry.Message,
	}
}

func makeEventInfo(entry *logrus.Entry) ecslogs.EventInfo {
	return ecslogs.EventInfo{
		Errors: makeErrors(entry.Data),
	}
}

func makeEventData(entry *logrus.Entry) ecslogs.EventData {
	data := make(ecslogs.EventData, len(entry.Data))

	for k, v := range entry.Data {
		switch k {
		case "msg", "level", "time":
		default:
			data[k] = v
		}
	}

	return data
}

func makeLevel(level logrus.Level) ecslogs.Level {
	switch level {
	case logrus.DebugLevel:
		return ecslogs.DEBUG

	case logrus.InfoLevel:
		return ecslogs.INFO

	case logrus.WarnLevel:
		return ecslogs.WARN

	case logrus.ErrorLevel:
		return ecslogs.ERROR

	case logrus.FatalLevel:
		return ecslogs.CRIT

	case logrus.PanicLevel:
		return ecslogs.ALERT

	default:
		return ecslogs.NONE
	}
}

func makeErrors(data logrus.Fields) (errors []ecslogs.EventError) {
	for k, v := range data {
		if err, ok := v.(error); ok {
			errors = append(errors, ecslogs.MakeEventError(err))
			delete(data, k)
		}
	}
	return
}
