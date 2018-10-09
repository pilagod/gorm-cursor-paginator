package paginator

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type fieldType string

const (
	fieldString fieldType = "STRING"
	fieldTime   fieldType = "TIME"
)

func convert(field interface{}) (result string) {
	switch field.(type) {
	case time.Time:
		result = fmt.Sprintf("%s?%s", field.(time.Time).UTC().Format(time.RFC3339Nano), fieldTime)
	default:
		result = fmt.Sprintf("%v?%s", field, fieldString)
	}
	return
}

func deconvert(field string) (result interface{}) {
	fieldTypeSepIndex := strings.LastIndex(field, "?")
	fieldType := fieldType(field[fieldTypeSepIndex+1:])
	field = field[:fieldTypeSepIndex]

	switch fieldType {
	case fieldTime:
		t, err := time.Parse(time.RFC3339Nano, field)

		if err != nil {
			t = time.Now().UTC()
		}
		result = t
	default:
		result = field
	}
	return
}

func flip(order order) order {
	if order == ASC {
		return DESC
	}
	return ASC
}

func reverse(v reflect.Value) reflect.Value {
	result := reflect.MakeSlice(v.Type(), 0, v.Cap())

	for i := v.Len() - 1; i >= 0; i-- {
		result = reflect.Append(result, v.Index(i))
	}
	return result
}
