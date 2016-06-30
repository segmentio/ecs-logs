package ecslogs

import (
	"errors"
	"io"
	"syscall"
	"testing"
)

type eventTest struct {
	e Event
	s string
}

func TestNewEvent(t *testing.T) {
	testEvents(t,
		eventTest{
			e: NewEvent(),
			s: `{}`,
		},
		eventTest{
			e: NewEvent(Tag{"answer", 42}, Tag{"question", "how are you?"}),
			s: `{"answer":42,"question":"how are you?"}`,
		},
	)
}

func TestEprintf(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprintf(""),
			s: `{"message":""}`,
		},
		eventTest{
			e: Eprintf("Hello World!"),
			s: `{"message":"Hello World!"}`,
		},
		eventTest{
			e: Eprintf("Answer: %d", 42),
			s: `{"message":"Answer: 42"}`,
		},
	)
}

func TestEprint(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprint(),
			s: `{"message":""}`,
		},
		eventTest{
			e: Eprint("Hello World!"),
			s: `{"message":"Hello World!"}`,
		},
		eventTest{
			e: Eprint("Answer: ", 42),
			s: `{"message":"Answer: 42"}`,
		},
	)
}

func TestEventErrors(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprint(io.EOF),
			s: `{"errors":[{"type":"*errors.errorString","error":"EOF"}],"message":"EOF"}`,
		},
		eventTest{
			e: Eprint(errors.New("A"), errors.New("B")),
			s: `{"errors":[{"type":"*errors.errorString","error":"A"},{"type":"*errors.errorString","error":"B"}],"message":"A B"}`,
		},
	)
}

func TestEventErrno(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprint(syscall.Errno(2)),
			s: `{"errno":2,"errors":[{"type":"syscall.Errno","error":"no such file or directory"}],"message":"no such file or directory"}`,
		},
	)
}

func TestEventFromMap(t *testing.T) {
	testEvents(t,
		eventTest{
			e: makeEventFrom(nil),
			s: `{}`,
		},
		eventTest{
			e: makeEventFrom(map[string]interface{}{}),
			s: `{}`,
		},
		eventTest{
			e: makeEventFrom(map[string]interface{}{"answer": 42}),
			s: `{"answer":42}`,
		},
		eventTest{
			e: makeEventFrom(map[int]interface{}{1: "A", 2: "B", 3: "C"}),
			s: `{"1":"A","2":"B","3":"C"}`,
		},
		eventTest{
			e: makeEventFrom(NewEvent()),
			s: `{}`,
		},
	)
}

func TestEventFromStruct(t *testing.T) {
	type A struct{}

	type B struct{ S string }

	type C struct {
		A int `json:"a"`
		B int `json:"b,omitempty"`
		C int `json:"-"`
	}

	testEvents(t,
		eventTest{
			e: makeEventFrom(A{}),
			s: `{}`,
		},
		eventTest{
			e: makeEventFrom(B{"Hello World!"}),
			s: `{"S":"Hello World!"}`,
		},
		eventTest{
			e: makeEventFrom(C{
				A: 42,
				B: 0,
				C: -1,
			}),
			s: `{"a":42}`,
		},
	)
}

func testEvents(t *testing.T, tests ...eventTest) {
	for _, test := range tests {
		if s := test.e.String(); s != test.s {
			t.Errorf("\n- expected: %s\n- found:    %s", s, test.s)
		}
	}
}
