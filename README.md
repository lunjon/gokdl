# GoKDL

A parser implementation for the [KDL](https://kdl.dev/) document language in Go.

## Example

The following code shows a minimal example of parsing a KDL document:

```go
package main

import (
    "log"
    "github.com/lunjon/gokdl"
)

func main() {
    bs := []byte(`
MyNode "string arg" myint=1234 awesome=true {
  child-node 
}      

// A node with arbitrary name (in quotes)
"Other node with much cooler name!" { Okay; }
`)

    doc, err := gokdl.Parse(bs)
    if err != nil {
        log.Fatal(err)
    }

    // Do something with doc ...
}
```

## API

The general API for module (including the types Doc, Node, Arg and Prop) is yet to be done.

Although it can be used, it is very rough.

## Implementation Status

- Comments
  - [x] Line
  - [x] Multiline
  - [x] Slash-dash
- [x] Node with children
- [x] Support arbitrary identifiers
- [x] Multiline nodes
- Number literals
  - [x] Integers
  - [x] Float
  - [x] Scientific notation
- Strings
  - [x] Regular strings (double quotes)
  - [x] Raw string literals
- [x] Type annotations
