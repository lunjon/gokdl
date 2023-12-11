package gokdl_test

import (
	"strings"
	"testing"

	"github.com/lunjon/gokdl"
)

func TestParseExample(t *testing.T) {
	doc := `
// Line comment

/*
multiline
	comment
*/

node "arg" prop=1

one; two; // Ignore this

nesting-testing /*ignore this as well*/ {
	child-1; child-?;

	child!THREE keyword="string" {
		nesting-should-work-here-as-well
	}
}

"Arbitrary name in quotes!"

integer-arg -1234
science-arg-a 1.78e12
science-arg-b 1.78e-3
science-arg-c 1.7883274

// Node on multiple lines
hello \
	1 2 3 \
	myProp="wow"
`

	r := strings.NewReader(doc)
	_, err := gokdl.Parse(r)
	if err != nil {
		t.Fatalf("expected no error but was: %s", err)
	}
}
