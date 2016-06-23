package ecslogs

import (
	"fmt"
	"strings"
)

type Level int

const (
	EMERGENCY Level = iota
	ALERT
	CRITICAL
	ERROR
	WARNING
	NOTICE
	INFO
	DEBUG
)

type ParseLevelError struct {
	Level string
}

func (e ParseLevelError) Error() string {
	return fmt.Sprintf("invalid message level %#v", e.Level)
}

func ParseLevel(s string) (lvl Level, err error) {
	switch strings.ToUpper(s) {
	case "EMERGENCY":
		lvl = EMERGENCY
	case "ALERT":
		lvl = ALERT
	case "CRITICAL":
		lvl = CRITICAL
	case "ERROR":
		lvl = ERROR
	case "WARNING":
		lvl = WARNING
	case "NOTICE":
		lvl = NOTICE
	case "INFO":
		lvl = INFO
	case "DEBUG":
		lvl = DEBUG
	default:
		err = ParseLevelError{s}
	}
	return
}

func (lvl Level) String() string {
	switch lvl {
	case EMERGENCY:
		return "EMERGENCY"
	case ALERT:
		return "ALERT"
	case CRITICAL:
		return "CRITICAL"
	case ERROR:
		return "ERROR"
	case WARNING:
		return "WARNING"
	case NOTICE:
		return "NOTICE"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return fmt.Sprintf("<%d>", lvl)
	}
}
