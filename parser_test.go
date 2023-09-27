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
		testname     string
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
		t.Run(test.testname, func(t *testing.T) {
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

func TestParserNodeIdentifierInvalid(t *testing.T) {
	tests := []struct {
		testname string
		ident    string
	}{
		{"integer", "1"},
		{"parenthesis", "a(b)c"},
		{"square brackets", "a[b]c"},
		{"equal", "a=c"},
		{"comma", "abcD,,Y"},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			parser := setup(test.ident)
			_, err := parser.parse()
			if err == nil {
				t.Fatal("expected error but was nil")
			}
		})
	}
}

func TestParserNodeArgs(t *testing.T) {
	// Arrange
	tests := []struct {
		testname         string
		body             string
		expectedNodeName string
		expectedArgValue any
	}{
		{"integer", "NodeName 1", "NodeName", 1},
		{"float", "NodeName 1.234", "NodeName", 1.234},
		{"string", `ohmy "my@value"`, "ohmy", "my@value"},
		{"null", `Nullify null`, "Nullify", nil},
		{"true", `MyBool true`, "MyBool", true},
		{"false", `MyBool false`, "MyBool", false},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			parser := setup(test.body)
			// Act
			doc, err := parser.parse()

			// Assert
			if err != nil {
				t.Fatal("expected error but was nil")
			}

			nodes := doc.Nodes()
			if len(nodes) != 1 {
				t.Fatal("expected nodes to be one")
			}
			node := nodes[0]

			if node.Name != test.expectedNodeName {
				t.Fatalf("expected node name to be %s but was %s", test.expectedNodeName, node.Name)
			}

			args := node.Args
			if len(args) != 1 {
				t.Fatal("expected node args to be one")
			}
			arg := args[0]

			if arg.Value != test.expectedArgValue {
				t.Fatalf("expected value to be %v but was %v", test.expectedArgValue, arg.Value)
			}
		})
	}
}

func TestParserNodeProp(t *testing.T) {
	// Arrange
	nodeName := "NodeName"
	tests := []struct {
		testname          string
		body              string
		expectedPropName  string
		expectedPropValue any
	}{
		{"integer value", "NodeName myprop=1", "myprop", 1},
		{"float value", "NodeName myprop=1.234", "myprop", 1.234},
		{"string value", `NodeName myprop="Hello, World!"`, "myprop", "Hello, World!"},
		{"string value - quoted name", `NodeName "hehe prop"="Hello, World!"`, "hehe prop", "Hello, World!"},
		{"null value", `NodeName myprop=null`, "myprop", nil},
		{"bool: true", `NodeName myprop=true`, "myprop", true},
		{"bool: false", `NodeName myprop=false`, "myprop", false},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			parser := setup(test.body)
			// Act
			doc, err := parser.parse()

			// Assert
			if err != nil {
				t.Fatalf("expected no error but was %s", err)
			}

			nodes := doc.Nodes()
			if len(nodes) != 1 {
				t.Fatal("expected nodes to be one")
			}

			node := nodes[0]
			if node.Name != nodeName {
				t.Fatalf("expected node name to be %s but was %s", nodeName, node.Name)
			}

			props := node.Props
			if len(props) != 1 {
				t.Fatal("expected node args to be one")
			}
			prop := props[0]

			if prop.Value != test.expectedPropValue {
				t.Fatalf("expected value to be %v but was %v", test.expectedPropValue, prop.Value)
			}

			if prop.Name != test.expectedPropName {
				t.Fatalf("expected name to be %v but was %v", test.expectedPropName, prop.Name)
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
