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
	// TODO: test with invalid strings
	tests := []struct {
		name string
		str  string
	}{
		{"integer - single digit", "1"},
		{"integer - multi digit", "12345"},
		{"integer - neg", "-12345"},
		{"float - dot", "1.1"},
		{"float - dot multi", "1.12345"},
		{"float - scientific (pos exp)", "1.123e12"},
		{"float - scientific (neg exp)", "1.123e-9"},
		{"float - scientific neg", "-11.123e9"},
	}

	for _, test := range tests {
		sc := setup(test.str)
		t.Run(test.name, func(t *testing.T) {
			token, _ := sc.Scan()
			if token != NUM {
				log.Fatalf("expected token to be %d but was %d", NUM, token)
			}
		})
	}
}

func setup(source string) *Scanner {
	r := strings.NewReader(source)
	return NewScanner(r)
}
