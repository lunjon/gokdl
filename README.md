# GoKDL

An implementation of the [kdl](https://kdl.dev/) document language in Go.

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
MyNode "string arg" {
  child-node withprop=1
}      

// A node with arbitrary name (in quotes)
"Other node with much cooler name!"
`)

    doc, err := gokdl.Parse(bs)
    if err != nil {
        log.Fatal(err)
    }

    // Do something with doc ...
}
```

## API

The generall API for module (including the types Doc, Node, Arg and Prop) is not done at all.

Although it can be used, it is very rough. So I'll work on that when the parsing is more mature.


## Implementation Status

- Comments
  - [x] Line
  - [x] Multiine
  - [ ] Slash-dash
- [x] Node with children
- [x] Support arbitrary identifiers
- [x] Multiline nodes
- Number literals
  - [x] Integers
  - [x] Float
  - [x] Scientific notation
  - [x] Negative numbers
- Strings
  - [x] Regular strings (double quotes)
  - [ ] Raw string literals
  - [ ] Raw string literals with `#`
- [ ] Type annotations
