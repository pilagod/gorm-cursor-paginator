package paginator

import (
	"reflect"
	"time"
)

type kind uint

const (
	kindInvalid kind = iota
	kindBool
	kindInt
	kindUint
	kindFloat
	kindString
	kindTime
)

func toKind(rt reflect.Type) kind {
	for rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	// Kind() treats time.Time as struct, so we need specific test for time.Time
	if rt.ConvertibleTo(reflect.TypeOf(time.Time{})) {
		return kindTime
	}
	switch rt.Kind() {
	case reflect.Bool:
		return kindBool
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return kindInt
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return kindUint
	case reflect.Float32, reflect.Float64:
		return kindFloat
	default:
		return kindString
	}
}
