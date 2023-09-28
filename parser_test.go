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
	tests := []struct {
		testname string
		body     string
	}{
		{"single line", "/* comment */"},
		{"single line - two comments", "/* comment */ /* another */"},
		{
			"multiple lines", `/*
comment
another
*/`,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			_ = setupAndParse(t, test.body)
		})
	}
}

func TestParserSlashdashCommentNode(t *testing.T) {
	doc := setupAndParse(t, `/-mynode`)
	nodes := doc.Nodes()
	if len(nodes) != 0 {
		t.Fatal("expected nodes to be empty")
	}
}

func TestParserSlashdashCommentArg(t *testing.T) {
	doc := setupAndParse(t, `Node.js /-"arg" 1`)
	nodes := doc.Nodes()
	if len(nodes) != 1 {
		t.Fatal("expected nodes to be 1")
	}

	args := nodes[0].Args
	if len(args) != 1 {
		t.Fatalf("expected args to be 1 but was %d", len(args))
	}

	if args[0].Value != 1 {
		t.Fatalf("expected arg value to be 1 but was %v", args[0].Value)
	}
}

func TestParserSlashdashCommentProp(t *testing.T) {
	doc := setupAndParse(t, `Node.js uncommented=true /-properly="arg" 1`)
	nodes := doc.Nodes()
	if len(nodes) != 1 {
		t.Fatal("expected nodes to be 1")
	}

	args := nodes[0].Args
	if len(args) != 1 {
		t.Fatalf("expected args to be 1 but was %d", len(args))
	}

	if args[0].Value != 1 {
		t.Fatalf("expected arg value to be 1 but was %v", args[0].Value)
	}

	props := nodes[0].Props
	if len(props) != 1 {
		t.Fatalf("expected props to be 1 but was %d", len(props))
	}

	if props[0].Value != true {
		t.Fatalf("expected prop value to be 'true' but was '%v'", props[0].Value)
	}
}

func TestParserSlashdashCommentChildren(t *testing.T) {
	doc := setupAndParse(t, `Node.js uncommented=true  1 /-{
	childNode
}`)
	nodes := doc.Nodes()
	if len(nodes) != 1 {
		t.Fatal("expected nodes to be 1")
	}

	children := nodes[0].Children
	if len(children) != 0 {
		t.Fatalf("expected children to be empty but was %d", len(children))
	}
}

func TestParserSlashdashCommentNestedChildren(t *testing.T) {
	doc := setupAndParse(t, `Node.js uncommented=true  1 {
	/-Ignored 1 2
	Exists true
}`)
	nodes := doc.Nodes()
	if len(nodes) != 1 {
		t.Fatal("expected nodes to be 1")
	}

	children := nodes[0].Children
	if len(children) != 1 {
		t.Fatalf("expected children to be 1 but was %d", len(children))
	}
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
		{"float1", "NodeName 1.234", "NodeName", 1.234},
		{"float2", "NodeName 1234.5678", "NodeName", 1234.5678},
		{"string1", `ohmy "my@value"`, "ohmy", "my@value"},
		{"string2", `ohmy "TODO: $1"`, "ohmy", "TODO: $1"},
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
				t.Fatalf("expected no error but was: %s", err)
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

func TestParserNodeArgsInvalid(t *testing.T) {
	// Arrange
	tests := []struct {
		testname string
		body     string
	}{
		{"integer followed by letter", "NodeName 1a"},
		{"bare identifier", "NodeName nodename"},
		{"unexpected slash", "NodeName /"},
		{"unexpected dot", "NodeName ."},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			// Act
			parser := setup(test.body)
			_, err := parser.parse()

			// Assert
			if err == nil {
				t.Fatal("expected error but was nil")
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

func TestParserNodePropInvalid(t *testing.T) {
	// Arrange
	tests := []struct {
		testname string
		body     string
	}{
		{"missing value", "NodeName myprop= "},
		{"identifier value", "NodeName myprop=identifier"},
		{"unterminated string", `NodeName myprop="opened`},
		{"parenthesis", `NodeName myprop=()`},
		{"misc1", `NodeName myprop=123a`},
		{"misc2", `NodeName myprop=1.23--`},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			// Act
			parser := setup(test.body)
			_, err := parser.parse()

			// Assert
			if err == nil {
				t.Fatal("expected error but was nil")
			}
		})
	}
}

func TestParserNodeChildren(t *testing.T) {
	tests := []struct {
		testname      string
		body          string
		expectedNodes int
	}{
		{"single line #1", "Parent { child1 }", 2},
		{"single line #2", "Parent { child1; child2 }", 3},
		{"single line #3", "Parent { child1; child2; }", 3},
		{"single line #4", "Parent { child1; /-child2; }", 2},
		{"single line #5", "Parent { /*child1*/ child2; }", 2},
		{
			"nested #1", `Parent {
	child1; child2
		}`,
			3,
		},
		{
			"nested #2", `Parent {
	child1;
	child-?
		}`,
			3,
		},
		{
			"nested #3", `Parent {
	child1 {}
	child-?
		}`,
			3,
		},
		{
			"nested #4", `Parent {
	child1 { child1-A }
	child-? }`,
			4,
		},
		{
			"nested #5", `Parent {
	child1 { child1-A }
	child-?

	deep-1 {
		deep-1-2 {
			/-deep-1-2-3-a
			deep-1-2-3-b
			deep-1-2-3-c
		}
	}
}`,
			8,
		},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			doc := setupAndParse(t, test.body)
			actual := totalChildren(doc)
			if actual != test.expectedNodes {
				t.Fatalf("expected %d total children but was %d", test.expectedNodes, actual)
			}
		})
	}

	doc := setupAndParse(t, `Parent { child-1; child2; child-3 }`)
	children := doc.nodes[0].Children
	if len(children) != 3 {
		t.Fatalf("expected to have 1 child but was %d", len(children))
	}
}

func TestParserNodeChildrenSingle(t *testing.T) {
	doc := setupAndParse(t, `Parent {
	child
}`)
	children := doc.nodes[0].Children
	if len(children) != 1 {
		t.Fatalf("expected to have 1 child but was %d", len(children))
	}
	if children[0].Name != "child" {
		t.Fatalf("expected to have name 'child' but was '%s'", children[0].Name)
	}
}

func TestParserNodeChildrenMultiple(t *testing.T) {
	doc := setupAndParse(t, `Parent {
	child-1; child2;
	child-3
}`)
	children := doc.nodes[0].Children
	if len(children) != 3 {
		t.Fatalf("expected to have 1 child but was %d", len(children))
	}
}

func TestParserNodeChildrenMultipleSameRow(t *testing.T) {
	doc := setupAndParse(t, `Parent { child-1; child2; child-3 }`)
	children := doc.nodes[0].Children
	if len(children) != 3 {
		t.Fatalf("expected to have 1 child but was %d", len(children))
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

func recNodeChildrenCount(node Node) int {
	if len(node.Children) == 0 {
		return 1
	}

	total := 1
	for _, ch := range node.Children {
		total += recNodeChildrenCount(ch)
	}
	return total
}

func totalChildren(doc Doc) int {
	total := 0
	for _, n := range doc.nodes {
		total += recNodeChildrenCount(n)
	}
	return total
}
