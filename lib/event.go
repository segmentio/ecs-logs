package ecslogs

import (
	"encoding/json"
	"fmt"
)

type Event map[string]interface{}

type Tag struct {
	Name  string
	Value interface{}
}

func NewEvent(level Level, tags ...Tag) Event {
	return make(Event).setTags(tags...).setLevel(level)
}

func Eprintf(level Level, format string, args ...interface{}) Event {
	return NewEvent(level).Printf(format, args...)
}

func Eprint(level Level, args ...interface{}) Event {
	return NewEvent(level).Print(args...)
}

func (e Event) String() string {
	return string(e.Bytes())
}

func (e Event) Bytes() []byte {
	b, _ := json.Marshal(e)
	return b
}

func (e Event) Printf(format string, args ...interface{}) Event {
	return e.setMessage(fmt.Sprintf(format, args...))
}

func (e Event) Print(args ...interface{}) Event {
	return e.setMessage(fmt.Sprint(args...))
}

func (e Event) Tag(name string, value interface{}) Event {
	e[name] = value
	return e
}

func (e Event) setLevel(level Level) Event {
	e["level"] = level
	return e
}

func (e Event) setSource(source string) Event {
	e["source"] = source
	return e
}

func (e Event) setMessage(msg string) Event {
	e["message"] = msg
	return e
}

func (e Event) setTags(tags ...Tag) Event {
	for _, t := range tags {
		e.Tag(t.Name, t.Value)
	}
	return e
}
