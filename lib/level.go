package ecslogs

import (
	"fmt"
	"strconv"
	"strings"
)

type Level int

const (
	NONE Level = iota
	EMERG
	ALERT
	CRIT
	ERROR
	WARN
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
	case "EMERG":
		lvl = EMERG
	case "ALERT":
		lvl = ALERT
	case "CRIT":
		lvl = CRIT
	case "ERROR":
		lvl = ERROR
	case "WARN":
		lvl = WARN
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
	case EMERG:
		return "EMERG"
	case ALERT:
		return "ALERT"
	case CRIT:
		return "CRIT"
	case ERROR:
		return "ERROR"
	case WARN:
		return "WARN"
	case NOTICE:
		return "NOTICE"
	case INFO:
		return "INFO"
	case DEBUG:
		return "DEBUG"
	default:
		return lvl.GoString()
	}
}

func (lvl Level) GoString() string {
	return "Level(" + strconv.Itoa(int(lvl)) + ")"
}

func (lvl Level) MarshalText() (b []byte, err error) {
	b = []byte(lvl.String())
	return
}

func (lvl *Level) UnmarshalText(b []byte) (err error) {
	*lvl, err = ParseLevel(string(b))
	return
}
