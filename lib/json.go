package ecslogs

import (
	"reflect"
	"strconv"
	"strings"
)

func jsonLen(v interface{}) (n int) {
	return jsonLenV(reflect.ValueOf(v))
}

func jsonLenV(v reflect.Value) (n int) {
	if v.IsValid() {
		switch t := v.Type(); t.Kind() {
		case reflect.Bool:
			return jsonLenBool(v.Bool())

		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return jsonLenInt(v.Int())

		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return jsonLenUint(v.Uint())

		case reflect.Float32, reflect.Float64:
			return jsonLenFloat(v.Float())

		case reflect.String:
			return jsonLenString(v.String())

		case reflect.Slice:
			if t.Elem().Kind() == reflect.Uint8 {
				return jsonLenBytes(v.Bytes())
			}
			return jsonLenArray(v)

		case reflect.Array:
			return jsonLenArray(v)

		case reflect.Map:
			return jsonLenMap(v)

		case reflect.Struct:
			return jsonLenStruct(t, v)

		case reflect.Ptr, reflect.Interface:
			if !v.IsNil() {
				return jsonLenV(v.Elem())
			}
		}
	}

	return jsonLenNull()
}

func jsonLenNull() (n int) {
	return 4
}

func jsonLenBool(v bool) (n int) {
	if v {
		return 4
	}
	return 5
}

func jsonLenInt(v int64) (n int) {
	if v == 0 {
		return 1
	}
	if v < 0 {
		n++
	}
	for v != 0 {
		v /= 10
		n++
	}
	return
}

func jsonLenUint(v uint64) (n int) {
	if v == 0 {
		return 1
	}
	for v != 0 {
		v /= 10
		n++
	}
	return
}

func jsonLenFloat(v float64) (n int) {
	var b [32]byte
	return len(strconv.AppendFloat(b[:0], v, 'g', -1, 64))
}

func jsonLenString(s string) (n int) {
	for _, c := range s {
		switch c {
		case '\n', '\t', '\r', '\v', '\b', '\f', '\\', '/', '"':
			n++
		}
	}
	return n + 2 + len(s)
}

func jsonLenBytes(b []byte) (n int) {
	// The standard json package uses base64 encoding for byte slices...
	n = len(b)
	return 2 + ((n * 4) / 3)
}

func jsonLenArray(v reflect.Value) (n int) {
	for i, j := 0, v.Len(); i != j; i++ {
		if i != 0 {
			n++
		}
		n += jsonLenV(v.Index(i))
	}
	return n + 2
}

func jsonLenMap(v reflect.Value) (n int) {
	for i, k := range v.MapKeys() {
		if i != 0 {
			n++
		}
		n += jsonLenV(k) + jsonLenV(v.MapIndex(k)) + 1
	}
	return n + 2
}

func jsonLenStruct(t reflect.Type, v reflect.Value) (n int) {
	for i, j := 0, v.NumField(); i != j; i++ {
		if name, omitempty, skip := parseJsonStructField(t.Field(i)); skip {
			continue
		} else if f := v.Field(i); omitempty && isEmptyValue(f) {
			continue
		} else {
			if n != 0 {
				n++
			}
			n += jsonLenString(name) + jsonLenV(f) + 1
		}
	}
	return n + 2
}

func parseJsonStructField(field reflect.StructField) (name string, omitempty bool, skip bool) {
	if name, omitempty, skip = parseJsonStructTag(field.Tag.Get("json")); len(name) == 0 {
		name = field.Name
	}
	return
}

func parseJsonStructTag(tag string) (name string, omitempty bool, skip bool) {
	name, tag = parseNextJsonTagToken(tag)
	token, _ := parseNextJsonTagToken(tag)
	skip = name == "-"
	omitempty = token == "omitempty"
	return
}

func parseNextJsonTagToken(tag string) (token string, next string) {
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
