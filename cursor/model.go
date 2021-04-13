package cursor

import "time"

type boolModel struct {
	Value    bool
	ValuePtr *bool
}

type intModel struct {
	Value    int
	ValuePtr *int
}

type uintModel struct {
	Value    uint
	ValuePtr *uint
}

type floatModel struct {
	Value    float64
	ValuePtr *float64
}

type stringModel struct {
	Value    string
	ValuePtr *string
}

type timeModel struct {
	Value    time.Time
	ValuePtr *time.Time
}

type structModel struct {
	Value    structValue
	ValuePtr *structValue
}

type structValue struct {
	Value []byte
}
