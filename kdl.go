package gokdl

import (
	"io"
)

// Parse the bytes into a KDL Document,
// returning an error if anything was invalid.
//
// The bytes must be valid unicode.
func Parse(r io.Reader) (Doc, error) {
	parser := newParser(r)
	return parser.parse()
}

// ValueType is the type name of the different
// primitive KDL types.
type ValueType string

const (
	TypeString ValueType = "string"
	TypeInt    ValueType = "int"
	TypeFloat  ValueType = "float"
	TypeBool   ValueType = "boolean"
	TypeNull   ValueType = "null"
)
