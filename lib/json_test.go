package ecslogs

import (
	"encoding/json"
	"io/ioutil"
	"testing"
	"time"
)

var (
	jsonLenTests = []interface{}{
		nil,

		true,
		false,

		0,
		1,
		42,
		-1,
		-42,
		0.1234,

		"",
		"Hello World!",
		"Hello\nWorld!",

		[]byte(""),
		[]byte("Hello World!"),

		json.Number("0"),
		json.Number("1.2345"),

		[]int{},
		[]int{1, 2, 3},
		[]string{"hello", "world"},
		[]interface{}{nil, true, 42, "hey!"},

		map[string]string{},
		map[string]int{"answer": 42},
		map[string]interface{}{
			"A": nil,
			"B": true,
			"C": 42,
			"D": "hey!",
		},

		struct{}{},
		struct{ Answer int }{42},
		struct {
			A int
			B int
			C int
		}{1, 2, 3},
		struct {
			Question string
			Answer   string
		}{"How are you?", "Well"},

		map[string]interface{}{
			"struct": struct {
				OK bool `json:",omitempty"`
			}{false},
			"what?": struct {
				List   []interface{}
				String string
			}{
				List:   []interface{}{1, 2, 3},
				String: "Hello World!",
			},
		},

		Event{
			Level:   DEBUG,
			Time:    time.Now(),
			Info:    EventInfo{Host: "localhost"},
			Data:    EventData{"hello": "world"},
			Message: "Hello World!",
		},
	}
)

func TestJsonLen(t *testing.T) {
	for _, test := range jsonLenTests {
		b, _ := json.Marshal(test)
		n := jsonLen(test)

		if n != len(b) {
			t.Errorf("%#v => %d != %d (%s)", test, n, len(b), string(b))
		}
	}
}

func BenchmarkJsonLen(b *testing.B) {
	for i := 0; i != b.N; i++ {
		for _, test := range jsonLenTests {
			jsonLen(test)
		}
	}
}

func BenchmarkJsonMarshal(b *testing.B) {
	for i := 0; i != b.N; i++ {
		for _, test := range jsonLenTests {
			json.Marshal(test)
		}
	}
}

func BenchmarkJsonMarshalDevNull(b *testing.B) {
	for i := 0; i != b.N; i++ {
		for _, test := range jsonLenTests {
			e := json.NewEncoder(ioutil.Discard)
			e.Encode(test)
		}
	}
}
