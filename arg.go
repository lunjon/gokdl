package gokdl

import (
	"fmt"
	"strconv"
)

type Arg struct {
	Value          any
	TypeAnnotation TypeAnnotation
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

func newArg(value any, ta TypeAnnotation) Arg {
	return Arg{
		Value:          value,
		TypeAnnotation: ta,
	}
}

func newStringArg(value, typeAnnot string) (Arg, error) {
	val, err := parseStringValue(value, typeAnnot)
	return Arg{
		Value:          val,
		TypeAnnotation: TypeAnnotation(typeAnnot),
	}, err
}

func newIntArg(value, typeAnnot string) (Arg, error) {
	val, err := parseIntValue(value, typeAnnot)
	return Arg{
		Value:          val,
		TypeAnnotation: TypeAnnotation(typeAnnot),
	}, err
}

func newFloatArg(value, typeAnnot string) (Arg, error) {
	val, err := parseFloatValue(value, typeAnnot)
	return Arg{
		Value:          val,
		TypeAnnotation: TypeAnnotation(typeAnnot),
	}, err
}
