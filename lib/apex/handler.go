package apex_ecslogs

import (
	"io"

	apex "github.com/apex/log"
	"github.com/segmentio/ecs-logs/lib"
)

func NewHandler(w io.Writer) apex.Handler {
	return NewHandlerWith(ecslogs.LoggerConfig{
		Output: ecslogs.NewLoggerOutput(w),
	})
}

func NewHandlerWith(config ecslogs.LoggerConfig) apex.Handler {
	logger := ecslogs.NewLoggerWith(config)
	return apex.HandlerFunc(func(entry *apex.Entry) (err error) {
		logger.Log(makeEvent(entry))
		return
	})
}

func makeEvent(entry *apex.Entry) ecslogs.Event {
	return ecslogs.Event{
		Level:   makeLevel(entry.Level),
		Info:    makeEventInfo(entry),
		Data:    makeEventData(entry),
		Time:    entry.Timestamp,
		Message: entry.Message,
	}
}

func makeEventInfo(entry *apex.Entry) ecslogs.EventInfo {
	return ecslogs.EventInfo{
		Errors: makeErrors(entry.Fields),
	}
}

func makeEventData(entry *apex.Entry) ecslogs.EventData {
	data := make(ecslogs.EventData, len(entry.Fields))

	for k, v := range entry.Fields {
		data[k] = v
	}

	return data
}

func makeLevel(level apex.Level) ecslogs.Level {
	switch level {
	case apex.DebugLevel:
		return ecslogs.DEBUG

	case apex.InfoLevel:
		return ecslogs.INFO

	case apex.WarnLevel:
		return ecslogs.WARN

	case apex.ErrorLevel:
		return ecslogs.ERROR

	case apex.FatalLevel:
		return ecslogs.CRIT

	default:
		return ecslogs.NONE
	}
}

func makeErrors(fields apex.Fields) (errors []ecslogs.EventError) {
	for k, v := range fields {
		if err, ok := v.(error); ok {
			errors = append(errors, ecslogs.MakeEventError(err))
			delete(fields, k)
		}
	}
	return
}
