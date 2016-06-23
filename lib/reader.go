package ecslogs

import (
	"encoding/json"
	"io"
)

type Reader interface {
	io.Closer

	ReadMessage() (Message, error)
}

func NewMessageDecoder(r io.Reader) Reader {
	return decoder{
		j: json.NewDecoder(r),
		r: r,
	}
}

type decoder struct {
	j *json.Decoder
	r io.Reader
}

func (d decoder) Close() (err error) {
	if c, ok := d.r.(io.Closer); ok {
		err = c.Close()
	}
	return
}

func (d decoder) ReadMessage() (msg Message, err error) {
	err = d.j.Decode(&msg)
	return
}
