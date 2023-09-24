package gokdl

import (
	"fmt"
	"unicode/utf8"
)

func Unmarshal(bs []byte) (Document, error) {
	if !utf8.Valid(bs) {
		return Document{}, fmt.Errorf("document must contain valid UTF-8")
	}

	parser := newParser(bs)
	return parser.parse()
}
