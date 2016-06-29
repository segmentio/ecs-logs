package ecslogs

import "encoding/json"

type Content struct {
	Raw   []byte
	Value interface{}
}

func (c Content) String() string {
	b, _ := c.MarshalJSON()
	return string(b)
}

func (c Content) MarshalJSON() (b []byte, err error) {
	if b = c.Raw; len(b) == 0 {
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
