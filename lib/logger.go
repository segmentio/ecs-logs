package ecslogs

import (
	"io"
	"os"
)

type Logger struct {
	level  Level
	output LoggerOutput
	depth  int
	event  Event
	caller func(int) (string, int, string, bool)
}

type LoggerConfig struct {
	Level  Level
	Output LoggerOutput
	Depth  int
	Event  Event
	Caller func(int) (string, int, string, bool)
}

type LoggerOutput interface {
	Send(Event)
}

type LoggerOutputFunc func(Event)

func (f LoggerOutputFunc) Send(e Event) { f(e) }

func NewLoggerOutput(w io.Writer) LoggerOutput {
	return LoggerOutputFunc(func(e Event) {
		w.Write(append(e.Bytes(), '\n'))
	})
}

func NewLogger(w io.Writer) *Logger {
	return NewLoggerWith(LoggerConfig{})
}

func NewLoggerWith(config LoggerConfig) *Logger {
	if config.Level == NONE {
		config.Level = DEBUG
	}

	if config.Output == nil {
		config.Output = NewLoggerOutput(os.Stdout)
	}

	if config.Caller == nil {
		config.Caller = defaultCaller
	}

	return &Logger{
		level:  config.Level,
		output: config.Output,
		depth:  config.Depth,
		caller: config.Caller,
		event:  config.Event.Copy(),
	}
}

func (log *Logger) Level() Level {
	return log.level
}

func (log *Logger) With(x interface{}) *Logger {
	return &Logger{
		level:  log.level,
		output: log.output,
		depth:  log.depth,
		caller: log.caller,
		event:  makeEventFrom(x).addEvent(log.event),
	}
}

func (log *Logger) Debugf(format string, args ...interface{}) {
	log.printf(1, DEBUG, format, args...)
}

func (log *Logger) Infof(format string, args ...interface{}) {
	log.printf(1, INFO, format, args...)
}

func (log *Logger) Noticef(format string, args ...interface{}) {
	log.printf(1, NOTICE, format, args...)
}

func (log *Logger) Warnf(format string, args ...interface{}) {
	log.printf(1, WARN, format, args...)
}

func (log *Logger) Errorf(format string, args ...interface{}) {
	log.printf(1, ERROR, format, args...)
}

func (log *Logger) Critf(format string, args ...interface{}) {
	log.printf(1, CRIT, format, args...)
}

func (log *Logger) Alertf(format string, args ...interface{}) {
	log.printf(1, ALERT, format, args...)
}

func (log *Logger) Emergf(format string, args ...interface{}) {
	log.printf(1, EMERG, format, args...)
}

func (log *Logger) Printf(level Level, format string, args ...interface{}) {
	log.printf(1, level, format, args...)
}

func (log *Logger) Debug(args ...interface{}) {
	log.print(1, DEBUG, args...)
}

func (log *Logger) Info(args ...interface{}) {
	log.print(1, INFO, args...)
}

func (log *Logger) Notice(args ...interface{}) {
	log.print(1, NOTICE, args...)
}

func (log *Logger) Warn(args ...interface{}) {
	log.print(1, WARN, args...)
}

func (log *Logger) Error(args ...interface{}) {
	log.print(1, ERROR, args...)
}

func (log *Logger) Crit(args ...interface{}) {
	log.print(1, CRIT, args...)
}

func (log *Logger) Alert(args ...interface{}) {
	log.print(1, ALERT, args...)
}

func (log *Logger) Emerg(args ...interface{}) {
	log.print(1, EMERG, args...)
}

func (log *Logger) Print(level Level, args ...interface{}) {
	log.print(1, level, args...)
}

func (log *Logger) Log(level Level, event Event) {
	log.log(1, level, event)
}

func (log *Logger) printf(depth int, level Level, format string, args ...interface{}) {
	if level <= log.level {
		log.log(depth+1, level, Eprintf(format, args...))
	}
}

func (log *Logger) print(depth int, level Level, args ...interface{}) {
	if level <= log.level {
		log.log(depth+1, level, Eprint(args...))
	}
}

func (log *Logger) log(depth int, level Level, event Event) {
	if level <= log.level {
		for k, v := range log.event {
			event[k] = v
		}

		if file, line, fn, ok := log.caller(log.depth + depth + 1); ok {
			event.setSource(MessageSource(file, line, fn))
		}

		event.setLevel(level)
		log.output.Send(event)
	}
}

var (
	defaultLogger = NewLoggerWith(LoggerConfig{
		Output: NewLoggerOutput(os.Stdout),
		Depth:  1,
		Caller: Caller,
	})
)

func With(v interface{}) *Logger {
	return defaultLogger.With(v)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Noticef(format string, args ...interface{}) {
	defaultLogger.Noticef(format, args...)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func Critf(format string, args ...interface{}) {
	defaultLogger.Critf(format, args...)
}

func Alertf(format string, args ...interface{}) {
	defaultLogger.Alertf(format, args...)
}

func Emergf(format string, args ...interface{}) {
	defaultLogger.Emergf(format, args...)
}

func Printf(level Level, format string, args ...interface{}) {
	defaultLogger.Printf(level, format, args...)
}

func Debug(args ...interface{}) {
	defaultLogger.Debug(args...)
}

func Info(args ...interface{}) {
	defaultLogger.Info(args...)
}

func Notice(args ...interface{}) {
	defaultLogger.Notice(args...)
}

func Warn(args ...interface{}) {
	defaultLogger.Warn(args...)
}

func Error(args ...interface{}) {
	defaultLogger.Error(args...)
}

func Crit(args ...interface{}) {
	defaultLogger.Crit(args...)
}

func Alert(args ...interface{}) {
	defaultLogger.Alert(args...)
}

func Emerg(args ...interface{}) {
	defaultLogger.Emerg(args...)
}

func Print(level Level, args ...interface{}) {
	defaultLogger.Print(level, args...)
}

func Log(level Level, event Event) {
	defaultLogger.Log(level, event)
}

func defaultCaller(depth int) (file string, line int, fn string, ok bool) {
	return
}
