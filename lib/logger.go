package ecslogs

import (
	"io"
	"os"
)

type Logger struct {
	output io.Writer
	depth  int
	caller func(int) (string, int, string, bool)
}

type LoggerConfig struct {
	Output io.Writer
	Depth  int

	// This field is used by unit tests to mock the caller function
	// and not depend on the line numbers in the tests sources.
	caller func(int) (string, int, string, bool)
}

func NewLogger(w io.Writer) *Logger {
	return NewLoggerWith(LoggerConfig{
		Output: w,
	})
}

func NewLoggerWith(config LoggerConfig) *Logger {
	if config.caller == nil {
		config.caller = caller
	}
	return &Logger{
		output: config.Output,
		depth:  config.Depth,
		caller: config.caller,
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

func (log *Logger) Log(event Event) {
	log.log(1, event)
}

func (log *Logger) printf(depth int, level Level, format string, args ...interface{}) {
	log.log(depth+1, Eprintf(level, format, args...))
}

func (log *Logger) print(depth int, level Level, args ...interface{}) {
	log.log(depth+1, Eprint(level, args...))
}

func (log *Logger) log(depth int, event Event) {
	if file, line, fn, ok := log.caller(log.depth + depth + 1); ok {
		event.setSource(MessageSource(file, line, fn))
	}
	log.output.Write(append(event.Bytes(), '\n'))
}

var (
	defaultLogger = NewLoggerWith(LoggerConfig{
		Output: os.Stdout,
		Depth:  1,
	})
)

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

func Log(event Event) {
	defaultLogger.Log(event)
}
