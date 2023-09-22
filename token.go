package gokdl

import "unicode"

var eof = rune(0)

type Token int

const (
	EOF Token = iota
	WS
	INVALID

	// Literals
	IDENT
	NUM    // Integers (for now)
	BOOL   // true | false
	STREAM // Contiguous string of non-whitespace characters

	// Special characters
	SEMICOLON         // ;
	CBRACK_OPEN       // {
	CBRACK_CLOSE      // }
	QUOTE             // "
	EQUAL             // =
	COMMENT_LINE      // //
	COMMENT_MUL_OPEN  // /*
	COMMENT_MUL_CLOSE // */
	COMMENT_SD        // /- (slash-dash)
	BACKSLASH         // \
	FORWSLASH         // /
	PAREN_OPEN        // (
	PAREN_CLOSE       // )
	GREAT             // >
	LESS              // <
	SBRACK_OPEN       // [
	SBRACK_CLOSE      // ]
	COMMA             // ,
	RAWSTR_OPEN       // r"
	RAWSTR_HASH_OPEN  // r#[...]"
	RAWSTR_HASH_CLOSE // "#[...]

	// Other characters
	CHAR
)

func IsInitialIdentToken(t Token) bool {
	return t == CHAR || t == QUOTE
}

func IsIdentifierToken(t Token) bool {
	switch t {
	case NUM, CHAR:
		return true
	default:
		return false
	}
}

func IsIdentifier(r rune) bool {
	return !nonIdents[r] && !unicode.IsSpace(r)
}

func init() {
	nonIdents = map[rune]bool{}
	nons := "\\/(){}<>;[]=,"
	for _, r := range nons {
		nonIdents[r] = true
	}

	hexRunes = map[rune]bool{}
	for _, r := range "0123456789abcdef" {
		hexRunes[r] = true
	}
}

var (
	nonIdents map[rune]bool
	hexRunes  map[rune]bool
)
