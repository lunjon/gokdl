package gokdl

import (
	"fmt"
	"strconv"
)

type Arg struct {
	Value     any
	valueType ValueType
}

func (a Arg) String() string {
	return fmt.Sprint(a.Value)
}

func newBoolArg(lit string) (Arg, error) {
	b, err := strconv.ParseBool(lit)
	return Arg{
		Value: b,
	}, err
}

func newArg(value any, t ValueType) Arg {
	return Arg{
		Value:     value,
		valueType: t,
	}
}
