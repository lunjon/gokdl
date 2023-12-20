# GoKDL

A parser implementation for the [KDL](https://kdl.dev/) document language in Go.

## Example

The following code shows a minimal example of parsing a KDL document:

```go
package main

import (
    "log"
    "strings"
    "github.com/lunjon/gokdl"
)

func main() {
    kdl := `
MyNode "string arg" myint=1234 awesome=true {
  child-node 
}      

// A node with arbitrary name (in quotes)
"Other node with much cooler name!" { Okay; }
`

    r := strings.NewReader(kdl)
    doc, err := gokdl.Parse(r)
    if err != nil {
        log.Fatal(err)
    }

    // Do something with doc ...
}
```

## API

Although the module can be used, and the API is still very rough,
I'm grateful for any feedback and suggestion regarding the API!
