package gokdl

import (
	"strconv"
)

type Arg struct {
	Value any
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
