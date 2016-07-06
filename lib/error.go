package lib

import "strings"

type ErrorList []error

func AppendError(err error, other error) error {
	if err == nil {
		return other
	}
	switch e := err.(type) {
	case ErrorList:
		return append(e, other)
	default:
		return ErrorList{err, other}
	}
}

func (err ErrorList) Error() string {
	s := make([]string, len(err))

	for i, e := range err {
		s[i] = e.Error()
	}

	return strings.Join(s, "\n")
}
