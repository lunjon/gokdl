package internal

import (
	"log"
	"strings"
	"testing"
)

func TestScannerScanWhitespace(t *testing.T) {
	tests := []struct {
		name string
		str  string
	}{
		{"empty", " "},
		{"newline", "\n"},
		{"multi newline", " \n\n"},
	}

	for _, test := range tests {
		sc := setup(test.str)
		t.Run(test.name, func(t *testing.T) {
			token, _ := sc.Scan()
			if token != WS {
				log.Fatalf("expected token to be %d but was %d", WS, token)
			}
		})
	}
}

func TestScannerScanNumbers(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		expectedToken Token
	}{
		{"integer - single digit", "1", NUM_INT},
		{"integer - multi digit", "12345", NUM_INT},
		{"integer - neg", "-12345", NUM_INT},
		{"integer - prefix", "+12345", NUM_INT},
		{"float - dot", "1.1", NUM_FLOAT},
		{"float - dot multi", "1.12345", NUM_FLOAT},
		{"float - scientific (pos exp)", "1.123e12", NUM_SCI},
		{"float - scientific (neg exp)", "1.123e-9", NUM_SCI},
		{"float - scientific neg", "-1.123e9", NUM_SCI},
	}

	for _, test := range tests {
		sc := setup(test.str)
		t.Run(test.name, func(t *testing.T) {
			token, _ := sc.Scan()
			if token != test.expectedToken {
				log.Fatalf("expected token to be %d but was %d", test.expectedToken, token)
			}
		})
	}
}

func setup(source string) *Scanner {
	r := strings.NewReader(source)
	return NewScanner(r)
}
