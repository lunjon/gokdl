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
	// fmt.Println("[scanner] CH:", string(ch))

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
	case '+':
		next := s.read()
		s.r.UnreadRune()

		if unicode.IsDigit(next) {
			s.r.UnreadRune()
			return s.scanNumber(false)
		}

		token = CHAR
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
	start := s.ScanWhile(unicode.IsDigit)
	next := s.read()
	if neg {
		start = "-" + start
	}

	if next == EOF_RUNE {
		return s.setAndReturn(NUM_INT, start)
	}

	comp := start + string(next)

	if strings.HasSuffix(comp, ".") {
		return s.scanFloat(comp)
	} else if comp == "0x" {
		return s.scanHex()
	} else if comp == "0o" {
		return s.scanOctal()
	} else if comp == "0b" {
		return s.scanBinary()
	}

	if next != '_' {
		if unicode.IsSpace(next) || !unicode.IsDigit(next) {
			s.r.UnreadRune()
			return s.setAndReturn(NUM_INT, start)
		}
	}

	// Read as integer
	s.r.UnreadRune()
	lit := s.ScanWhile(func(r rune) bool {
		return unicode.IsDigit(r) || r == '_'
	})

	return s.setAndReturn(NUM_INT, strings.ReplaceAll(start+lit, "_", ""))
}

func (s *Scanner) scanFloat(start string) (Token, string) {
	// Try scientific notation: 1.234e-42
	if len(strings.TrimPrefix(start, "-")) == 2 {
		numsAfterDot := s.ScanWhile(unicode.IsDigit)
		if numsAfterDot == "" {
			return s.setAndReturn(CHARS, start)
		}

		tokenAfterNums, sAfterNums := s.Scan()

		if tokenAfterNums == CHAR && sAfterNums == "e" {
			next, ch := s.Scan()
			var exp string

			if ch == "-" {
				exp = s.ScanWhile(unicode.IsDigit)
				exp = "-" + exp
			} else if next == NUM_INT {
				exp = ch
			} else {
				return CHARS, start + numsAfterDot + sAfterNums + ch
			}

			num := fmt.Sprintf("%s%se%s", start, numsAfterDot, exp)
			return s.setAndReturn(NUM_SCI, num)
		} else if tokenAfterNums == NUM_INT {
			num := start + numsAfterDot + sAfterNums
			return s.setAndReturn(NUM_FLOAT, num)
		} else if tokenAfterNums == WS || tokenAfterNums == EOF {
			s.Unread()
			return s.setAndReturn(NUM_FLOAT, start+numsAfterDot)
		}

	}

	numsAfterDot := s.ScanWhile(unicode.IsDigit)
	if numsAfterDot == "" {
		return s.setAndReturn(CHARS, start)
	}

	return s.setAndReturn(NUM_FLOAT, start+numsAfterDot)
}

func (s *Scanner) scanBinary() (Token, string) {
	// Read binary
	lit := s.ScanWhile(func(r rune) bool {
		return r == '0' || r == '1' || r == '_'
	})
	lit = strings.ReplaceAll(lit, "_", "")

	n, err := strconv.ParseInt(lit, 2, 64)
	if err != nil {
		return s.setAndReturn(CHARS, "0b"+lit)
	}

	return s.setAndReturn(NUM_INT, fmt.Sprint(n))
}

func (s *Scanner) scanOctal() (Token, string) {
	// Read binary
	lit := s.ScanWhile(func(r rune) bool {
		return ('0' <= r && r <= '7') || r == '_'
	})
	lit = strings.ReplaceAll(lit, "_", "")

	n, err := strconv.ParseInt(lit, 8, 64)
	if err != nil {
		return s.setAndReturn(CHARS, "0o"+lit)
	}

	return s.setAndReturn(NUM_INT, fmt.Sprint(n))
}

func (s *Scanner) scanHex() (Token, string) {
	// Read hexadecimal: 0xdeadbeef
	lit := s.ScanWhile(func(r rune) bool {
		return hexRunes[r] || r == '_'
	})
	lit = strings.ReplaceAll(lit, "_", "")

	n, err := strconv.ParseInt(lit, 16, 64)
	if err != nil {
		return s.setAndReturn(CHARS, "0x"+lit)
	}

	return s.setAndReturn(NUM_INT, fmt.Sprint(n))
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
