package internal

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// Scanner represents a lexical Scanner.
type Scanner struct {
	r   *bufio.Reader
	eof bool
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{r: bufio.NewReader(r)}
}

func (s *Scanner) ScanLine() {
	if s.eof {
		return
	}
	_, _ = s.r.ReadBytes('\n')
}

// scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	ch := s.read()

	if unicode.IsSpace(ch) {
		s.Unread()
		return s.ScanWhitespace()
	} else if unicode.IsDigit(ch) {
		s.Unread()
		return s.ScanNumber(false)
	}

	switch ch {
	case EOF_RUNE:
		return EOF, ""
	case '"':
		return QUOTE, string(ch)
	case '=':
		return EQUAL, string(ch)
	case '-':
		next := s.read()
		s.Unread()

		if unicode.IsDigit(next) {
			s.Unread()
			return s.ScanNumber(true)
		}

		return HYPHEN, string(ch)
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
		if ch := s.read(); ch == EOF_RUNE {
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

// ScanNumber tries to scan a number in any of the supported formats.
// Use `neg` to indicate that the number was prefixed with a hyphen.
func (s *Scanner) ScanNumber(neg bool) (Token, string) {
	first := s.read()
	second := s.read()
	final := func(s string) string {
		if neg {
			return "-" + s
		}
		return s
	}

	if second == EOF_RUNE {
		s.Unread()
		return NUM, final(string(first))
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

			num := fmt.Sprintf("%s.%se%s", string(first), numsAfterDot, exp)
			return NUM, final(num)
		} else if tokenAfterNums == WS || tokenAfterNums == EOF {
			s.Unread()
			num := fmt.Sprintf("%s.%s", string(first), numsAfterDot)
			return NUM, final(num)
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

		return NUM, final(fmt.Sprint(n))
	} else if second == 'b' {
		// Read binary
		lit := s.ScanWhile(func(r rune) bool {
			return r == '0' || r == '1'
		})

		n, err := strconv.ParseInt(lit, 2, 64)
		if err != nil {
			return INVALID, ""
		}

		return NUM, final(fmt.Sprint(n))
	} else if unicode.IsSpace(second) {
		s.Unread()
		return NUM, final(string(first))
	}

	// Read as integer
	s.Unread()
	lit := s.ScanWhile(func(r rune) bool {
		return unicode.IsDigit(r) || r == '_'
	})

	return NUM, final(strings.ReplaceAll(lit, "_", ""))
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

// scanLetters consumes the current rune and all contiguous ident runes.
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
		return EOF_RUNE
	}
	return r
}

// unread places the previously read rune back on the reader.
func (s *Scanner) Unread() {
	_ = s.r.UnreadRune()
}
