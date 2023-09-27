package gokdl

import (
	"fmt"
	"strconv"
)

type ArgType string

const (
	TypeString ArgType = "string"
	TypeInt    ArgType = "int"
	TypeFloat  ArgType = "float"
	TypeBool   ArgType = "boolean"
	TypeNull   ArgType = "null"
)

type Arg struct {
	Value    any
	typeInfo string
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

func newArg(value any, t ArgType) Arg {
	return Arg{
		Value:    value,
		typeInfo: "",
	}
}
