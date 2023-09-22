package gokdl

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"unicode"
)

// Scanner represents a lexical scanner.
type Scanner struct {
	r   *bufio.Reader
	eof bool
}

// NewScanner returns a new instance of Scanner.
func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) ScanLine() {
	if s.eof {
		return
	}
	_, _ = s.r.ReadBytes('\n')
}

// Scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if unicode.IsSpace(ch) {
		s.Unread()
		return s.ScanWhitespace()
	} else if unicode.IsDigit(ch) {
		s.Unread()
		return s.ScanNumber()
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
		s.Unread()
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
			s.Unread()
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

func (s *Scanner) ScanWhile(pred func(rune) bool) string {
	var buf bytes.Buffer
	buf.WriteRune(s.read())

	for {
		if ch := s.read(); ch == eof {
			break
		} else if !pred(ch) {
			s.Unread()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return buf.String()
}

func (s *Scanner) ScanNumber() (Token, string) {
	first := s.read()
	second := s.read()

	if second == eof {
		s.Unread()
		return NUM, string(first)
	}

	if second == '.' {
		// Read scientific notation: 1.234e-42
		numsAfterDot := s.ScanWhile(unicode.IsDigit)
		if numsAfterDot == "" {
			return INVALID, ""
		}

		tokenAfterNums, sAfterNums := s.Scan()

		if tokenAfterNums == CHAR && sAfterNums == "e" {
			next, ch := s.Scan()
			var exp string

			if ch == "-" {
				exp = s.ScanWhile(unicode.IsDigit)
				exp = "-" + exp
			} else if next == NUM {
				exp = ch
			} else {
				return INVALID, ""
			}

			return NUM, fmt.Sprintf("%s.%se%s", string(first), numsAfterDot, exp)
		} else if tokenAfterNums == WS {
			s.Unread()
			return NUM, fmt.Sprintf("%s.%s", string(first), numsAfterDot)
		}

		return INVALID, ""
	} else if second == 'x' {
		// Read hexadecimal: 0xdeadbeef
		lit := s.ScanWhile(func(r rune) bool {
			return hexRunes[r]
		})

		n, err := strconv.ParseInt(lit, 16, 64)
		if err != nil {
			return INVALID, ""
		}

		return NUM, fmt.Sprint(n)
	} else if second == 'b' {
		// Read binary
		lit := s.ScanWhile(func(r rune) bool {
			return r == '0' || r == '1'
		})

		n, err := strconv.ParseInt(lit, 2, 64)
		if err != nil {
			return INVALID, ""
		}

		return NUM, fmt.Sprint(n)
	} else if unicode.IsSpace(second) {
		s.Unread()
		return NUM, string(first)
	}

	s.Unread()
	lit := s.ScanWhile(unicode.IsDigit)
	return NUM, lit
}

// Scan while whitespace only.
func (s *Scanner) ScanWhitespace() (Token, string) {
	lit := s.ScanWhile(unicode.IsSpace)
	return WS, lit
}

// Scan while non-whitespice.
func (s *Scanner) ScanNonWhitespace() string {
	return s.ScanWhile(func(r rune) bool {
		return !unicode.IsSpace(r)
	})
}

// ScanLetters consumes the current rune and all contiguous ident runes.
func (s *Scanner) ScanLetters() (Token, string) {
	pred := func(r rune) bool {
		return unicode.IsLetter(r) || r == '_'
	}

	lit := s.ScanWhile(pred)
	return IDENT, lit
}

func (s *Scanner) ScanBareIdent() string {
	return s.ScanWhile(IsIdentifier)
}

// Read the next rune from the reader.
// Returns the rune(0) if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.eof = true
		return eof
	}
	return r
}

// Unread places the previously read rune back on the reader.
func (s *Scanner) Unread() {
	_ = s.r.UnreadRune()
}
