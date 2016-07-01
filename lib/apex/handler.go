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
	// Extract the FuncInfo field from the logger configuration, that way the
	// default logic for looking up the caller information will not be executed
	// and we can provide one that is compatible with the apex/log package.
	funcInfo := config.FuncInfo
	config.FuncInfo = nil

	logger := ecslogs.NewLoggerWith(config)

	if funcInfo == nil {
		return apex.HandlerFunc(func(entry *apex.Entry) error {
			return logger.Log(makeEvent(entry, ""))
		})
	}

	return apex.HandlerFunc(func(entry *apex.Entry) error {
		var source string

		if pc, ok := ecslogs.GuessCaller(config.Depth, config.Depth+10, "github.com/segmentio/ecs-logs"); ok {
			if info, ok := funcInfo(pc); ok {
				source = info.String()
			}
		}

		return logger.Log(makeEvent(entry, source))
	})
}

func makeEvent(entry *apex.Entry, source string) ecslogs.Event {
	return ecslogs.Event{
		Level:   makeLevel(entry.Level),
		Info:    makeEventInfo(entry, source),
		Data:    makeEventData(entry),
		Time:    entry.Timestamp,
		Message: entry.Message,
	}
}

func makeEventInfo(entry *apex.Entry, source string) ecslogs.EventInfo {
	return ecslogs.EventInfo{
		Source: source,
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
