package gokdl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

// scanner represents a lexical scanner.
type scanner struct {
	r   *bufio.Reader
	eof bool
}

func newScanner(r io.Reader) *scanner {
	return &scanner{r: bufio.NewReader(r)}
}

func (s *scanner) scanLine() {
	if s.eof {
		return
	}
	_, _ = s.r.ReadBytes('\n')
}

// scan returns the next token and literal value.
func (s *scanner) scan() (tok Token, lit string) {
	ch := s.read()

	if unicode.IsSpace(ch) {
		s.unread()
		return s.scanWhitespace()
	} else if unicode.IsDigit(ch) {
		s.unread()
		return s.scanNumber()
	}

	switch ch {
	case eof:
		return EOF, ""
	case '"':
		return QUOTE, string(ch)
	case '=':
		return EQUAL, string(ch)
	case '*':
		next := s.read()
		if next == '/' {
			return COMMENT_MUL_CLOSE, "*/"
		}
		s.unread()
		return CHAR, string(ch)
	case '/':
		next := s.read()
		switch next {
		case '/':
			return COMMENT_LINE, "//"
		case '*':
			return COMMENT_MUL_OPEN, "/*"
		case '-':
			return COMMENT_SD, "/-"
		default:
			s.unread()
			return CHAR, string(ch)
		}
	case ';':
		return SEMICOLON, string(ch)
	case '{':
		return CBRACK_OPEN, string(ch)
	case '}':
		return CBRACK_CLOSE, string(ch)
	case '[':
		return SBRACK_OPEN, string(ch)
	case ']':
		return SBRACK_CLOSE, string(ch)
	case '<':
		return LESS, string(ch)
	case '>':
		return GREAT, string(ch)
	case ',':
		return COMMA, string(ch)
	case '(':
		return PAREN_OPEN, string(ch)
	case ')':
		return PAREN_CLOSE, string(ch)
	case '\\':
		return BACKSLASH, string(ch)
	default:
		return CHAR, string(ch)
	}
}

func (s *scanner) scanWhile(pred func(rune) bool) string {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !pred(ch) {
			s.unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return buf.String()
}

func (s *scanner) scanNumber() (Token, string) {
	first := s.read()
	second := s.read()

	if second == eof {
		s.unread()
		return NUM, string(first)
	}

	if second == '.' {
		// Read scientific notation: 1.234e-42
		numsAfterDot := s.scanWhile(unicode.IsDigit)
		if numsAfterDot == "" {
			return INVALID, ""
		}

		tokenAfterNums, sAfterNums := s.scan()

		if tokenAfterNums == CHAR && sAfterNums == "e" {
			next, ch := s.scan()
			var exp string

			if ch == "-" {
				exp = s.scanWhile(unicode.IsDigit)
				exp = "-" + exp
			} else if next == NUM {
				exp = ch
			} else {
				return INVALID, ""
			}

			return NUM, fmt.Sprintf("%s.%se%s", string(first), numsAfterDot, exp)
		} else if tokenAfterNums == WS {
			s.unread()
			return NUM, fmt.Sprintf("%s.%s", string(first), numsAfterDot)
		}

		return INVALID, ""
	} else if second == 'x' {
		// Read hexadecimal: 0xdeadbeef
		lit := s.scanWhile(func(r rune) bool {
			return hexRunes[r]
		})

		n, err := strconv.ParseInt(lit, 16, 64)
		if err != nil {
			return INVALID, ""
		}

		return NUM, fmt.Sprint(n)
	} else if second == 'b' {
		// Read binary
		lit := s.scanWhile(func(r rune) bool {
			return r == '0' || r == '1'
		})

		n, err := strconv.ParseInt(lit, 2, 64)
		if err != nil {
			return INVALID, ""
		}

		return NUM, fmt.Sprint(n)
	} else if unicode.IsSpace(second) {
		s.unread()
		return NUM, string(first)
	}

	s.unread()
	lit := s.scanWhile(unicode.IsDigit)
	return NUM, lit
}

// Scan while whitespace only.
func (s *scanner) scanWhitespace() (Token, string) {
	lit := s.scanWhile(unicode.IsSpace)
	return WS, lit
}

// Scan while non-whitespice.
func (s *scanner) scanNonWhitespace() string {
	return s.scanWhile(func(r rune) bool {
		return !unicode.IsSpace(r)
	})
}

// scanLetters consumes the current rune and all contiguous ident runes.
func (s *scanner) scanLetters() (Token, string) {
	pred := func(r rune) bool {
		return unicode.IsLetter(r) || r == '_'
	}

	lit := s.scanWhile(pred)
	return IDENT, lit
}

func (s *scanner) scanBareIdent() string {
	return s.scanWhile(IsIdentifier)
}

// Read the next rune from the reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.eof = true
		return eof
	}
	return r
}

// unread places the previously read rune back on the reader.
func (s *scanner) unread() {
	_ = s.r.UnreadRune()
}
