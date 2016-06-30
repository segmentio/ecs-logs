package ecslogs

import "testing"

type eventTest struct {
	e Event
	s string
}

func TestNewEvent(t *testing.T) {
	testEvents(t,
		eventTest{
			e: NewEvent(DEBUG),
			s: `{"level":"DEBUG"}`,
		},
		eventTest{
			e: NewEvent(INFO, Tag{"answer", 42}, Tag{"question", "how are you?"}),
			s: `{"answer":42,"level":"INFO","question":"how are you?"}`,
		},
	)
}

func TestEprintf(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprintf(DEBUG, ""),
			s: `{"level":"DEBUG","message":""}`,
		},
		eventTest{
			e: Eprintf(INFO, "Hello World!"),
			s: `{"level":"INFO","message":"Hello World!"}`,
		},
		eventTest{
			e: Eprintf(NOTICE, "Answer: %d", 42),
			s: `{"level":"NOTICE","message":"Answer: 42"}`,
		},
	)
}

func TestEprint(t *testing.T) {
	testEvents(t,
		eventTest{
			e: Eprint(DEBUG),
			s: `{"level":"DEBUG","message":""}`,
		},
		eventTest{
			e: Eprint(INFO, "Hello World!"),
			s: `{"level":"INFO","message":"Hello World!"}`,
		},
		eventTest{
			e: Eprint(NOTICE, "Answer: ", 42),
			s: `{"level":"NOTICE","message":"Answer: 42"}`,
		},
	)
}

func testEvents(t *testing.T, tests ...eventTest) {
	for _, test := range tests {
		if s := test.e.String(); s != test.s {
			t.Errorf("%s != %s", s, test.s)
		}
	}
}
