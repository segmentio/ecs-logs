package ecslogs

import "reflect"

func levelValue(v interface{}) (lvl Level, ok bool) {
	var err error

	if s, x := stringValue(v); !x {
		return
	} else if lvl, err = ParseLevel(s); err != nil {
		return
	}

	ok = true
	return
}

func timeValue(v interface{}) (t Timestamp, ok bool) {
	var err error

	if s, x := stringValue(v); !x {
		return
	} else if t, err = ParseTimestamp(s); err != nil {
		return
	}

	ok = true
	return
}

func intValue(v interface{}) (n int, ok bool) {
	switch x := reflect.ValueOf(v); x.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n = int(x.Int())

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n = int(x.Uint())

	case reflect.Float32, reflect.Float64:
		n = int(x.Float())

	default:
		return
	}

	ok = true
	return
}

func stringValue(v interface{}) (s string, ok bool) {
	switch x := v.(type) {
	case string:
		s = x

	case []byte:
		s = string(x)

	default:
		switch x := reflect.ValueOf(v); x.Kind() {
		case reflect.String:
			s = x.String()

		default:
			return
		}
	}

	ok = true
	return
}
