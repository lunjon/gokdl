package gokdl

import (
	"fmt"
	"strconv"
)

type Arg struct {
	Value any
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

func newArg(lit string) Arg {
	return Arg{
		Value: lit,
	}
}
