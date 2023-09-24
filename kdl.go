package gokdl

import (
	"fmt"
	"unicode/utf8"
)

// Parse the bytes into a KDL Document,
// returning an error if anything was invalid.
//
// The bytes must be valid unicode.
func Parse(bs []byte) (Doc, error) {
	if !utf8.Valid(bs) {
		return Doc{}, fmt.Errorf("document must contain valid UTF-8")
	}
	parser := newParser(bs)
	return parser.parse()
}
