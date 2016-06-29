package ecslogs

import (
	"encoding/json"
	"strconv"
)

type Content struct {
	Raw   []byte
	Value map[string]interface{}
}

func MakeContent(s string) (c Content) {
	c.Raw = []byte(strconv.Quote(s))
	return
}

func (c Content) String() string {
	b, _ := c.MarshalJSON()
	return string(b)
}

func (c Content) MarshalJSON() (b []byte, err error) {
	if c.Value == nil {
		b = c.Raw
	} else {
		b, err = json.Marshal(c.Value)
	}
	return
}

func (c *Content) UnmarshalJSON(b []byte) (err error) {
	c.Raw = make([]byte, len(b))
	copy(c.Raw, b)

	if json.Unmarshal(b, &c.Value) != nil {
		c.Value = nil
	}

	return
}
