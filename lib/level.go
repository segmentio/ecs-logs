package ecslogs

import (
	"encoding/json"
	"fmt"
	"strconv"
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
		if v, e := strconv.ParseInt(s, 10, 64); e != nil {
			err = ParseLevelError{s}
		} else {
			lvl = Level(v)
		}
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
		return strconv.Itoa(int(lvl))
	}
}

func (lvl Level) MarshalJSON() (b []byte, err error) {
	return json.Marshal(lvl.String())
}

func (lvl *Level) UnmarshalJSON(b []byte) (err error) {
	var v int
	var s string

	if err = json.Unmarshal(b, &v); err == nil {
		*lvl = Level(v)
		return
	}

	if err = json.Unmarshal(b, &s); err == nil {
		*lvl, err = ParseLevel(s)
		return
	}

	return
}
