package main

import (
	"fmt"
	"log"

	"github.com/lunjon/gokdl"
)

func main() {
	bs := []byte(`
// Line comment
/*
multiline
	comment
*/

node "arg" prop=1

one; two; // Ignore this

withchild /*ignore this as well*/ {
	child-1; child-?;
	child!THREE keyword="string" {
		seriousNesting
	}
}

"testing !!! this works, yay!"

integer-arg 1234
science-arg-a 1.78e12
science-arg-b 1.78e-3
science-arg-c 1.7883274

// Node on multiple lines
hello \
	1 2 3 \
	property="wowk"
`)
	doc, err := gokdl.Unmarshal(bs)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(doc.String())
}
