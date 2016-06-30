package ecslogs

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"syscall"
)

type Event map[string]interface{}

type Tag struct {
	Name  string
	Value interface{}
}

const (
	eventLevelKey   = "level"
	eventSourceKey  = "source"
	eventMessageKey = "message"
	eventErrorsKey  = "errors"
	eventErrnoKey   = "errno"
)

func NewEvent(tags ...Tag) Event {
	return make(Event, len(tags)).setTags(tags...)
}

func Eprintf(format string, args ...interface{}) Event {
	return NewEvent().Printf(format, args...)
}

func Eprint(args ...interface{}) Event {
	return NewEvent().Print(args...)
}

func (e Event) String() string {
	return string(e.Bytes())
}

func (e Event) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e Event) Printf(format string, args ...interface{}) Event {
	e.setValues(args...)
	return e.setMessage(fmt.Sprintf(format, args...))
}

func (e Event) Print(args ...interface{}) Event {
	e.setValues(args...)
	return e.setMessage(fmt.Sprint(args...))
}

func (e Event) Tag(name string, value interface{}) Event {
	e[name] = value
	return e
}

func (e Event) Level() Level {
	switch v := e[eventLevelKey].(type) {
	case Level:
		return v
	}
	return NONE
}

func (e Event) Copy() Event {
	return make(Event, len(e)).addEvent(e)
}

func (e Event) setLevel(level Level) Event {
	e[eventLevelKey] = level
	return e
}

func (e Event) setSource(source string) Event {
	e[eventSourceKey] = source
	return e
}

func (e Event) setMessage(msg string) Event {
	e[eventMessageKey] = msg
	return e
}

func (e Event) setTags(tags ...Tag) Event {
	for _, t := range tags {
		e.Tag(t.Name, t.Value)
	}
	return e
}

func (e Event) setValues(args ...interface{}) Event {
	for _, a := range args {
		e.setValue(a)
	}
	return e
}

func (e Event) setValue(arg interface{}) Event {
	switch v := arg.(type) {
	case syscall.Errno:
		return e.addError(v).addErrno(v)
	case error:
		return e.addError(v)
	}
	return e
}

func (e Event) addEvent(c Event) Event {
	for k, v := range c {
		e[k] = v
	}
	return e
}

func (e Event) addError(err error) Event {
	if errors := e[eventErrorsKey]; errors == nil {
		e[eventErrorsKey] = []eventError{makeEventError(err)}
	} else {
		e[eventErrorsKey] = append(errors.([]eventError), makeEventError(err))
	}
	return e
}

func (e Event) addErrno(errno syscall.Errno) Event {
	e[eventErrnoKey] = int(errno)
	return e
}

type eventError struct {
	Type  string `json:"type"`
	Error string `json:"error"`
}

func makeEventError(err error) eventError {
	return eventError{
		Type:  reflect.TypeOf(err).String(),
		Error: err.Error(),
	}
}

func makeEventFrom(x interface{}) Event {
	if x != nil {
		switch v := x.(type) {
		case Event:
			return v.Copy()

		case map[string]interface{}:
			return Event(v).Copy()
		}

		switch t, v := reflect.TypeOf(x), reflect.ValueOf(x); t.Kind() {
		case reflect.Struct:
			return makeEventFromStruct(t, v)

		case reflect.Map:
			return makeEventFromMap(t, v)

		case reflect.Ptr:
			if !v.IsNil() {
				return makeEventFrom(v.Elem().Interface())
			}
		}
	}
	return NewEvent()
}

func makeEventFromStruct(t reflect.Type, v reflect.Value) (e Event) {
	e = NewEvent()

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

	return
}

func makeEventFromMap(t reflect.Type, v reflect.Value) (e Event) {
	e = NewEvent()

	if t.Key().Kind() == reflect.String {
		for _, k := range v.MapKeys() {
			e[k.String()] = v.MapIndex(k).Interface()
		}
	} else {
		for _, k := range v.MapKeys() {
			e[fmt.Sprint(k.Interface())] = v.MapIndex(k).Interface()
		}
	}

	return
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
