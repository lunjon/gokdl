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

type previous struct {
	token Token
	lit   string
}

// Scanner represents a lexical Scanner.
type Scanner struct {
	r   *bufio.Reader
	eof bool
	// State used in unread.
	prev *previous // Set from last when Unread was called
	last previous
}

func NewScanner(r io.Reader) *Scanner {
	return &Scanner{
		r: bufio.NewReader(r),
	}
}

func (s *Scanner) ScanLine() {
	if s.eof {
		return
	}
	_, _ = s.r.ReadBytes('\n')
}

// scan returns the next token and literal value.
func (s *Scanner) Scan() (tok Token, lit string) {
	if s.eof {
		return EOF, ""
	}

	if s.prev != nil {
		token := s.prev.token
		lit := s.prev.lit
		s.prev = nil
		return token, lit
	}

	ch := s.read()

	if unicode.IsSpace(ch) {
		s.r.UnreadRune()
		return s.ScanWhitespace()
	} else if unicode.IsDigit(ch) {
		s.r.UnreadRune()
		return s.scanNumber(false)
	}

	var token Token
	var str string
	switch ch {
	case EOF_RUNE:
		s.eof = true
		token = EOF
	case '"':
		token = QUOTE
		str = string(ch)
	case '=':
		token = EQUAL
		str = string(ch)
	case '-':
		next := s.read()
		s.r.UnreadRune()

		if unicode.IsDigit(next) {
			s.r.UnreadRune()
			return s.scanNumber(true)
		}

		token = HYPHEN
		str = string(ch)
	case '*':
		next := s.read()
		if next == '/' {
			token = COMMENT_MUL_CLOSE
			str = "*/"
		} else {
			s.r.UnreadRune()
			token = CHAR
			str = string(ch)
		}
	case '/':
		next := s.read()
		switch next {
		case '/':
			token = COMMENT_LINE
			str = "//"
		case '*':
			token = COMMENT_MUL_OPEN
			str = "/*"
		case '-':
			token = COMMENT_SD
			str = "/-"
		default:
			s.r.UnreadRune()
			return CHAR, string(ch)
		}
	case ';':
		token = SEMICOLON
		str = string(ch)
	case '{':
		token = CBRACK_OPEN
		str = string(ch)
	case '}':
		token = CBRACK_CLOSE
		str = string(ch)
	case '[':
		token = SBRACK_OPEN
		str = string(ch)
	case ']':
		token = SBRACK_CLOSE
		str = string(ch)
	case '<':
		token = LESS
		str = string(ch)
	case '>':
		token = GREAT
		str = string(ch)
	case ',':
		token = COMMA
		str = string(ch)
	case '(':
		token = PAREN_OPEN
		str = string(ch)
	case ')':
		token = PAREN_CLOSE
		str = string(ch)
	case '\\':
		token = BACKSLASH
		str = string(ch)
	default:
		token = CHAR
		str = string(ch)
	}

	return s.setAndReturn(token, str)
}

func (s *Scanner) ScanWhile(pred func(rune) bool) string {
	var buf bytes.Buffer
	if s.prev != nil {
		buf.WriteString(s.prev.lit)
		s.prev = nil
	}

	for {
		ch := s.read()
		if ch == EOF_RUNE {
			break
		} else if !pred(ch) {
			s.r.UnreadRune()
			break
		} else {
			buf.WriteRune(ch)
		}
	}

	return buf.String()
}

// scanNumber tries to scan a number in any of the supported formats.
// Use `neg` to indicate that the number was prefixed with a hyphen.
func (s *Scanner) scanNumber(neg bool) (Token, string) {
	first := s.read()
	second := s.read()
	final := func(s string) string {
		if neg {
			return "-" + s
		}
		return s
	}

	if second == EOF_RUNE {
		s.r.UnreadRune()
		return s.setAndReturn(NUM, final(string(first)))
	}

	if second == '.' {
		// Read scientific notation: 1.234e-42
		numsAfterDot := s.ScanWhile(unicode.IsDigit)
		if numsAfterDot == "" {
			return s.setAndReturn(INVALID, "")
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
			return s.setAndReturn(NUM, final(num))
		} else if tokenAfterNums == WS || tokenAfterNums == EOF {
			s.Unread()
			num := fmt.Sprintf("%s.%s", string(first), numsAfterDot)
			return s.setAndReturn(NUM, final(num))
		}

		return s.setAndReturn(INVALID, "")
	} else if second == 'x' {
		// Read hexadecimal: 0xdeadbeef
		lit := s.ScanWhile(func(r rune) bool {
			return hexRunes[r]
		})

		n, err := strconv.ParseInt(lit, 16, 64)
		if err != nil {
			return s.setAndReturn(INVALID, "")
		}

		return s.setAndReturn(NUM, final(fmt.Sprint(n)))
	} else if second == 'b' {
		// Read binary
		lit := s.ScanWhile(func(r rune) bool {
			return r == '0' || r == '1'
		})

		n, err := strconv.ParseInt(lit, 2, 64)
		if err != nil {
			return s.setAndReturn(INVALID, "")
		}

		return s.setAndReturn(NUM, final(fmt.Sprint(n)))
	} else if unicode.IsSpace(second) {
		s.r.UnreadRune()
		return s.setAndReturn(NUM, final(string(first)))
	}

	// Read as integer
	s.r.UnreadRune()
	lit := s.ScanWhile(func(r rune) bool {
		return unicode.IsDigit(r) || r == '_'
	})

	return s.setAndReturn(NUM, final(strings.ReplaceAll(lit, "_", "")))
}

// Scan while whitespace only.
func (s *Scanner) ScanWhitespace() (Token, string) {
	lit := s.ScanWhile(unicode.IsSpace)
	return s.setAndReturn(WS, lit)
}

// scanLetters consumes the current rune and all contiguous ident runes.
func (s *Scanner) ScanLetters() (Token, string) {
	pred := func(r rune) bool {
		return unicode.IsLetter(r) || r == '_'
	}

	lit := s.ScanWhile(pred)
	return s.setAndReturn(IDENT, lit)
}

func (s *Scanner) ScanBareIdent() string {
	lit := s.ScanWhile(IsIdentifier)
	s.setAndReturn(IDENT, lit)
	return lit
}

// Read the next rune from the reader.
// Returns `eof` if an error occurs (or io.EOF is returned).
func (s *Scanner) read() rune {
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.eof = true
		return EOF_RUNE
	}
	return r
}

func (s *Scanner) setAndReturn(t Token, lit string) (Token, string) {
	s.last = previous{token: t, lit: lit}
	return t, lit
}

func (s *Scanner) Unread() {
	s.prev = &s.last
}
