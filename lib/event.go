package ecslogs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"syscall"
)

type EventError struct {
	Type  string `json:"type,omitempty"`
	Error string `json:"error,omitempty"`
	Errno int    `json:"errno,omitempty"`
}

func MakeEventError(err error) EventError {
	return EventError{
		Type:  reflect.TypeOf(err).String(),
		Error: err.Error(),
	}
}

func MakeEventErrno(err syscall.Errno) EventError {
	return EventError{
		Type:  reflect.TypeOf(err).String(),
		Error: err.Error(),
		Errno: int(err),
	}
}

type EventInfo struct {
	Level  Level        `json:"level,omitempty"`
	Host   string       `json:"host,omitempty"`
	Source string       `json:"source,omitempty"`
	ID     string       `json:"id,omitempty"`
	PID    int          `json:"pid,omitempty"`
	UID    int          `json:"uid,omitempty"`
	GID    int          `json:"gid,omitempty"`
	Time   Timestamp    `json:"time,omitempty"`
	Errors []EventError `json:"errors,omitempty"`
}

func (c EventInfo) Bytes() []byte {
	b, _ := json.Marshal(c)
	return b
}

func (c EventInfo) String() string {
	return string(c.Bytes())
}

type EventData map[string]interface{}

func (c EventData) Bytes() []byte {
	b, _ := json.Marshal(c)
	return b
}

func (c EventData) String() string {
	return string(c.Bytes())
}

type Event struct {
	Info EventInfo `json:"info"`
	Data EventData `json:"data"`
}

func Eprintf(level Level, format string, args ...interface{}) Event {
	return MakeEvent(level, sprintf(format, args...), args...)
}

func Eprint(level Level, args ...interface{}) Event {
	return MakeEvent(level, sprint(args...), args...)
}

func MakeEvent(level Level, message string, values ...interface{}) Event {
	var errors []EventError

	for _, val := range values {
		switch v := val.(type) {
		case syscall.Errno:
			errors = append(errors, MakeEventErrno(v))
		case error:
			errors = append(errors, MakeEventError(v))
		}
	}

	return Event{
		Info: EventInfo{Level: level, Errors: errors},
		Data: EventData{"message": message},
	}
}

func (c Event) Bytes() []byte {
	b, _ := json.Marshal(c)
	return b
}

func (c Event) String() string {
	return string(c.Bytes())
}

func copyEventData(data ...EventData) EventData {
	copy := EventData{}

	for _, d := range data {
		for k, v := range d {
			copy[k] = v
		}
	}

	return copy
}

func makeEventData(x interface{}) EventData {
	if x != nil {
		switch v := x.(type) {
		case EventData:
			return copyEventData(v)

		case map[string]interface{}:
			return copyEventData(v)
		}

		switch t, v := reflect.TypeOf(x), reflect.ValueOf(x); t.Kind() {
		case reflect.Struct:
			return makeEventDataFromStruct(t, v)

		case reflect.Map:
			return makeEventDataFromMap(t, v)

		case reflect.Ptr:
			if !v.IsNil() {
				return makeEventData(v.Elem().Interface())
			}
		}
	}
	return EventData{}
}

func makeEventDataFromStruct(t reflect.Type, v reflect.Value) EventData {
	e := EventData{}

	for i, n := 0, t.NumField(); i != n; i++ {
		if ft, fv := t.Field(i), v.Field(i); ft.Anonymous {
			continue
		} else if name, omitempty, skip := parseStructField(ft); skip {
			continue
		} else if omitempty && isEmptyValue(fv) {
			continue
		} else {
			e[name] = fv.Interface()
		}
	}

	return e
}

func makeEventDataFromMap(t reflect.Type, v reflect.Value) EventData {
	e := EventData{}

	if t.Key().Kind() == reflect.String {
		for _, k := range v.MapKeys() {
			e[k.String()] = v.MapIndex(k).Interface()
		}
	} else {
		for _, k := range v.MapKeys() {
			e[fmt.Sprint(k.Interface())] = v.MapIndex(k).Interface()
		}
	}

	return e
}

func parseStructField(field reflect.StructField) (name string, omitempty bool, skip bool) {
	if name, omitempty, skip = parseStructTag(field.Tag.Get("json")); len(name) == 0 {
		name = field.Name
	}
	return
}

func parseStructTag(tag string) (name string, omitempty bool, skip bool) {
	name, tag = parseNextTagToken(tag)
	token, _ := parseNextTagToken(tag)
	skip = name == "-"
	omitempty = token == "omitempty"
	return
}

func parseNextTagToken(tag string) (token string, next string) {
	if split := strings.IndexByte(tag, ','); split < 0 {
		token = tag
	} else {
		token, next = tag[:split], tag[split+1:]
	}
	return
}

// Copied from https://golang.org/src/encoding/json/encode.go?h=isEmpty#L282
func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}

func sprint(args ...interface{}) string {
	s := fmt.Sprintln(args...)
	return s[:len(s)-1]
}
