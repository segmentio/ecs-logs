package ecslogs

import (
	"encoding/json"
	"io"
	"os"
	"runtime"
)

type Logger interface {
	Log(Event) error
}

type LoggerFunc func(Event) error

func (f LoggerFunc) Log(e Event) error {
	return f(e)
}

type LoggerConfig struct {
	Output   io.Writer
	Depth    int
	Data     EventData
	FuncInfo func(uintptr) (FuncInfo, bool)
}

func NewLogger(config LoggerConfig) Logger {
	if config.Output == nil {
		config.Output = os.Stdout
	}

	enc := json.NewEncoder(config.Output)

	return LoggerFunc(func(event Event) error {
		if event.Data == nil {
			event.Data = EventData{}
		}

		if len(config.Data) != 0 {
			// Copy the default events set on the logger, but do not overwrite
			// keys that already exist.
			for k, v := range config.Data {
				if _, x := event.Data[k]; !x {
					event.Data[k] = v
				}
			}
		}

		if config.FuncInfo != nil {
			if pc, _, _, ok := runtime.Caller(config.Depth + 2); ok {
				if info, ok := config.FuncInfo(pc); ok {
					event.Info.Source = info.String()
				}
			}
		}

		return enc.Encode(event)
	})
}
