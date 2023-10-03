package internal

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
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
				t.Fatalf("expected token to be %d but was %d", WS, token)
			}
		})
	}
}

func TestScannerScanNumbers(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		expectedToken Token
		expectedLit   string
	}{
		{"integer - single digit", "1", NUM_INT, "1"},
		{"integer - multi digit", "12345", NUM_INT, "12345"},
		{"integer - neg", "-12345", NUM_INT, "-12345"},
		{"integer - prefix", "+12345", NUM_INT, "12345"},
		{"integer - underscore", "10_000", NUM_INT, "10000"},
		{"float - dot", "1.1", NUM_FLOAT, "1.1"},
		{"float - dot multi", "1.12345", NUM_FLOAT, "1.12345"},
		{"float - scientific (pos exp)", "1.123e12", NUM_SCI, "1.123e12"},
		{"float - scientific (neg exp)", "1.123e-9", NUM_SCI, "1.123e-9"},
		{"float - scientific neg", "-1.123e9", NUM_SCI, "-1.123e9"},
		{"binary", "0b0101", NUM_INT, "5"},
		{"binary - underscore", "0b01_01", NUM_INT, "5"},
		{"octal", "0o010463", NUM_INT, "4403"},
		{"octal - underscore", "0o0104_63", NUM_INT, "4403"},
		{"hex", "0xabc123", NUM_INT, "11256099"},
		{"hex - underscore", "0xabc_123", NUM_INT, "11256099"},
	}

	for _, test := range tests {
		sc := setup(test.str)
		t.Run(test.name, func(t *testing.T) {
			token, lit := sc.Scan()
			require.Equal(t, test.expectedToken, token)
			require.Equal(t, test.expectedLit, lit)
		})
	}
}

func TestScannerScanRawString(t *testing.T) {
	tests := []struct {
		name          string
		str           string
		expectedToken Token
		expectedLit   string
	}{
		{"no raw string", "r", CHAR, "r"},
		{"raw string", `r"`, RAWSTR_OPEN, `r"`},
		{"raw string hash 1", `r#"`, RAWSTR_HASH_OPEN, `r#"`},
		{"raw string hash 2", `r##"`, RAWSTR_HASH_OPEN, `r##"`},
		{"other", `r##`, CHAR, `r##`},
		{"quote end 1", `"`, QUOTE, `"`},
		{"quote end 2", `"#`, RAWSTR_HASH_CLOSE, `"#`},
		{"quote end 3", `"##`, RAWSTR_HASH_CLOSE, `"##`},
	}

	for _, test := range tests {
		sc := setup(test.str)
		t.Run(test.name, func(t *testing.T) {
			token, lit := sc.Scan()
			require.Equal(t, test.expectedToken, token)
			require.Equal(t, test.expectedLit, lit)
		})
	}
}

func setup(source string) *Scanner {
	r := strings.NewReader(source)
	return NewScanner(r)
}
