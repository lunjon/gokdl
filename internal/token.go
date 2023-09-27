package internal

import (
	"unicode"
)

var EOF_RUNE = rune(0)

type Token int

const (
	EOF Token = iota
	WS
	INVALID

	// Literals
	IDENT
	NUM_INT   // Integer
	NUM_FLOAT // Float
	NUM_SCI   // Scientific notation
	BOOL      // true | false

	// Special characters
	SEMICOLON         // ;
	CBRACK_OPEN       // {
	CBRACK_CLOSE      // }
	QUOTE             // "
	EQUAL             // =
	HYPHEN            // -
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
	return t == CHAR || t == QUOTE || t == HYPHEN
}

func IsIdentifierToken(t Token) bool {
	switch t {
	case NUM_INT, CHAR:
		return true
	default:
		return false
	}
}

func IsIdentifier(r rune) bool {
	return !nonIdents[r] && !unicode.IsSpace(r)
}

func IsAnyOf(t Token, ts ...Token) bool {
	for _, ot := range ts {
		if t == ot {
			return true
		}
	}
	return false
}

func ContainsNonIdent(s string) bool {
	for _, ch := range s {
		if nonIdents[ch] {
			return true
		}
	}
	return false
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
	// Runes that are not valid in identifiers
	nonIdents map[rune]bool
	hexRunes  map[rune]bool
)
