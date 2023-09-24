package gokdl

import (
	"strings"
)

type Document struct {
	nodes []Node
}

func (d Document) String() string {
	nodes := []string{}
	for _, node := range d.nodes {
		nsdfnsd := stringRecNode("", node)
		nodes = append(nodes, nsdfnsd)
	}

	return strings.Join(nodes, "\n")
}
