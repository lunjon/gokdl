package gokdl

import (
	"io"
	"log"
	"testing"
)

func TestParserLineComment(t *testing.T) {
	_ = setupAndParse(t, `// First line
// Second line
// Thirdline`)
}

func TestParserMultilineComment(t *testing.T) {
	_ = setupAndParse(t, `/*
First line
/Second line
Thirdline. */`)
}

func TestParserValidNodeIdentifier(t *testing.T) {
	tests := []struct {
		testName     string
		doc          string
		expectedName string
	}{
		{"lower case letters", "node", "node"},
		{"snake case", "node_name", "node_name"},
		{"end with number", "node_name123", "node_name123"},
		{"arbitrary characters #1", `-this_actually::WORKS?`, "-this_actually::WORKS?"},
		{"quoted named", `"Node Name?"`, "Node Name?"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			doc := setupAndParse(t, test.doc)
			if len(doc.nodes) != 1 {
				t.Fatalf("expected nodes to have length 1 but was %d", len(doc.nodes))
			}

			name := doc.nodes[0].Name
			if name != test.expectedName {
				t.Fatalf("expected node to have name %s but was %s", test.expectedName, name)
			}
		})
	}
}

func TestParserInvalidNodeIdentifier(t *testing.T) {
	tests := []struct {
		testName string
		ident    string
	}{
		{"integer", "1"},
		{"parenthesis", "a(b)c"},
		{"square brackets", "a[b]c"},
		{"equal", "a=c"},
		{"comma", "abcD,,Y"},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			parser := setup(test.ident)
			_, err := parser.parse()
			if err == nil {
				t.Fatal("expected error but was nil")
			}
		})
	}
}

func setup(doc string) *parser {
	logger := log.New(io.Discard, "", 0)
	return newParser(logger, []byte(doc))
}

func setupAndParse(t *testing.T, doc string) Doc {
	p := setup(doc)
	d, err := p.parse()
	if err != nil {
		t.Fatalf("expected no error but was: %s", err)
	}
	return d
}
