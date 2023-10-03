package gokdl

import (
	"io"
	"log"
	"os"

	// "os"
	"testing"

	"github.com/stretchr/testify/require"
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
	// Arrange & Act
	doc := setupAndParse(t, "Node.js /-\"arg\" 1")

	// Assert
	nodes := doc.Nodes()
	require.Len(t, nodes, 1)
	args := nodes[0].Args
	require.Len(t, args, 1)
	require.Equal(t, int64(1), args[0].Value)
}

func TestParserSlashdashCommentProp(t *testing.T) {
	doc := setupAndParse(t, "Node.js uncommented=true /-properly=\"arg\" 1")
	nodes := doc.Nodes()
	require.Len(t, nodes, 1)

	args := nodes[0].Args
	require.Len(t, args, 1)
	require.Equal(t, int64(1), args[0].Value)

	props := nodes[0].Props
	require.Len(t, props, 1)

	require.Equal(t, true, props[0].Value)
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
		{"arbitrary characters #1", "-this_actually::WORKS?", "-this_actually::WORKS?"},
		{"quoted named", "\"Node Name?\"", "Node Name?"},
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
	nodeName := "node"
	tests := []struct {
		testname         string
		body             string
		expectedArgValue any
	}{
		{"integer", "node 1", int64(1)},
		{"integer with underscore", "node 1_0_0", int64(100)},
		{"float1", "node 1.234", 1.234},
		{"float2", "node 1234.5678", 1234.5678},
		{"string1", "node \"my@value\"", "my@value"},
		{"string2", `node "TODO: $1"`, "TODO: $1"},
		{"string3", `node "log.Printf(\"$1\")"`, `log.Printf("$1")`},
		{"string4", `node "block{
	$1
}"`, `block{
	$1
}`},
		{"rawstring1", `node r"h\e\l\l"`, `h\e\l\l`},
		{"rawstringhash1", `node r#"h\e\l\l"#`, `h\e\l\l`},
		{"rawstringhash2", `node r##"h\e\l\l"##`, `h\e\l\l`},
		{"rawstringhash3", `node r##"he"ll"##`, `he"ll`},
		{"rawstringhash4", `node r##"he#ll"##`, `he#ll`},
		{"null", "node null", nil},
		{"true", "node true", true},
		{"false", "node false", false},
		{"hex - small caps", "node 0x1aaeff", int64(1748735)},
		{"hex - mixed caps", "node 0x1AAeff", int64(1748735)},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			// Act
			parser := setup(test.body)
			doc, err := parser.parse()

			// Assert
			require.NoError(t, err)

			nodes := doc.Nodes()
			require.Len(t, nodes, 1)
			node := nodes[0]
			require.Equal(t, nodeName, node.Name)

			require.Len(t, node.Args, 1)
			arg := node.Args[0]

			require.Equal(t, test.expectedArgValue, arg.Value)
			require.Equal(t, TypeAnnotation(""), arg.TypeAnnotation)
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
		{"unterminated string", `NodeName ".`},
		{"invalid termination of raw string 1", `NodeName r".`},
		{"invalid termination of raw string 2", `NodeName r##"."#`},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			// Act
			parser := setup(test.body)
			_, err := parser.parse()

			// Assert
			require.Error(t, err)
		})
	}
}

func TestParserNodeArgsTypeAnnotationsInvalid(t *testing.T) {
	// Arrange
	tests := []struct {
		testname string
		body     string
	}{
		{"type annotation for invalid literal: null", "NodeName (u8)null"},
		{"type annotation for invalid literal: true", "NodeName (u8)true"},
		{"type annotation for invalid literal: false", "NodeName (u8)false"},
		{"u8 for type string", `NodeName (u8)"value"`},
		{"uncloses paranthesis", `NodeName (string"value"`},
		{"integer for type float", "NodeName (u16)12.456"},
		{"float for type integer", "NodeName (f64)12"},
		{"negative for unsigned integer", "NodeName (u64)-12"},
		{"overflow for u8", "NodeName (u8)1024"},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			// Act
			parser := setup(test.body)
			_, err := parser.parse()

			// Assert
			require.Error(t, err)
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
		{"integer value", "NodeName myprop=1", "myprop", int64(1)},
		{"float value", "NodeName myprop=1.234", "myprop", 1.234},
		{"string value", "NodeName myprop=\"Hello, World!\"", "myprop", "Hello, World!"},
		{"string value - quoted name", "NodeName \"hehe prop\"=\"Hello, World!\"", "hehe prop", "Hello, World!"},
		{"null value", "NodeName myprop=null", "myprop", nil},
		{"bool: true", "NodeName myprop=true", "myprop", true},
		{"bool: false", "NodeName myprop=false", "myprop", false},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			parser := setup(test.body)
			// Act
			doc, err := parser.parse()

			// Assert
			require.NoError(t, err)

			nodes := doc.Nodes()
			require.Len(t, nodes, 1)

			node := nodes[0]
			require.Equal(t, nodeName, node.Name)

			props := node.Props
			require.Len(t, props, 1)
			prop := props[0]

			require.Equal(t, test.expectedPropName, prop.Name)
			require.Equal(t, test.expectedPropValue, prop.Value)
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

func TestParserNodePropTypeAnnotation(t *testing.T) {
	// Arrange
	nodeName := "NodeName"
	propName := "myprop"
	tests := []struct {
		testname               string
		body                   string
		expectedValue          any
		expectedTypeAnnot      TypeAnnotation
		expectedValueTypeAnnot TypeAnnotation
	}{
		{"integer value - type annotation on arg", "NodeName myprop=(i64)1", int64(1), noTypeAnnot, I64},
		{"integer value - type annotation on prop", "NodeName (author)myprop=1", int64(1), TypeAnnotation("author"), noTypeAnnot},
		{"integer value - type annotation on prop and arg", "NodeName (author)myprop=(i64)1", int64(1), TypeAnnotation("author"), I64},
	}

	for _, test := range tests {
		t.Run(test.testname, func(t *testing.T) {
			parser := setup(test.body)
			// Act
			doc, err := parser.parse()

			// Assert
			require.NoError(t, err)

			nodes := doc.Nodes()
			require.Len(t, nodes, 1)

			node := nodes[0]
			require.Equal(t, nodeName, node.Name)

			props := node.Props
			require.Len(t, props, 1)
			prop := props[0]

			require.Equal(t, propName, prop.Name)
			require.Equal(t, test.expectedValue, prop.Value)
			require.Equal(t, test.expectedTypeAnnot, prop.TypeAnnot)
			require.Equal(t, test.expectedValueTypeAnnot, prop.ValueTypeAnnot)
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

func TestParserStringsEscaped(t *testing.T) {
	// Arrange
	filename := "testdata/escaped.kdl"
	bs, err := os.ReadFile(filename)
	require.NoError(t, err)
	logger := log.New(io.Discard, "", 0)
	parser := newParser(logger, bs)

	// Act
	doc, err := parser.parse()

	//Assert
	require.NoError(t, err)
	nodes := doc.Nodes()
	require.Equal(t, "\t", nodes[0].Args[0].Value)
	require.Equal(t, "\u00CA", nodes[1].Args[0].Value)
	require.Equal(t, "Ê", nodes[1].Args[0].Value)
	require.Equal(t, `"`, nodes[2].Args[0].Value)
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
