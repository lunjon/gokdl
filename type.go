package gokdl

import (
	"fmt"
	"strconv"
)

type TypeAnnotation string

func (t TypeAnnotation) String() string {
	return string(t)
}

const (
	noTypeAnnot                = ""
	I8          TypeAnnotation = "i8"
	I16         TypeAnnotation = "i16"
	I32         TypeAnnotation = "i32"
	I64         TypeAnnotation = "i64"
	U8          TypeAnnotation = "u8"
	U16         TypeAnnotation = "u16"
	U32         TypeAnnotation = "u32"
	U64         TypeAnnotation = "u64"
	F32         TypeAnnotation = "f32"
	F64         TypeAnnotation = "f64"
)

var (
	numberTypeAnnotation = map[string]TypeAnnotation{
		I8.String():  I8,
		I16.String(): I16,
		I32.String(): I32,
		I64.String(): I64,
		U8.String():  U8,
		U16.String(): U16,
		U32.String(): U32,
		U64.String(): U64,
		F32.String(): F32,
		F64.String(): F64,
	}
)

func init() {
	nums := []TypeAnnotation{
		I8,
		I16,
		I32,
		I64,
		U8,
		U16,
		U32,
		U64,
		F32,
		F64,
	}

	for _, n := range nums {
		numberTypeAnnotation[n.String()] = n
	}
}

func parseStringValue(value, typeAnnot string) (string, error) {
	if _, isNum := numberTypeAnnotation[typeAnnot]; isNum {
		return "", fmt.Errorf("invalid type annotation for type string: %s", typeAnnot)
	}
	return value, nil
}

func parseIntValue(value, typeAnnot string) (any, error) {
	var bitsize int
	var unsigned bool

	switch TypeAnnotation(typeAnnot) {
	case noTypeAnnot:
		bitsize = 64
	case I8:
		bitsize = 8
	case I16:
		bitsize = 16
	case I32:
		bitsize = 32
	case I64:
		bitsize = 64
	case U8:
		bitsize = 8
		unsigned = true
	case U16:
		bitsize = 16
		unsigned = true
	case U32:
		bitsize = 32
		unsigned = true
	case U64:
		bitsize = 64
		unsigned = true
	default:
		return value, fmt.Errorf("invalid type annotation for integer: %s", typeAnnot)
	}

	if unsigned {
		return strconv.ParseUint(value, 10, bitsize)
	} else {
		return strconv.ParseInt(value, 10, bitsize)
	}
}

func parseFloatValue(value, typeAnnot string) (any, error) {
	var bitsize int

	switch TypeAnnotation(typeAnnot) {
	case F32:
		bitsize = 32
	case noTypeAnnot, F64:
		bitsize = 64
	default:
		return value, fmt.Errorf("invalid type annotation for integer: %s", typeAnnot)
	}

	return strconv.ParseFloat(value, bitsize)
}
