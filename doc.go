package gokdl

import (
	"strings"
)

type Doc struct {
	nodes []Node
}

func (d Doc) Nodes() []Node {
	return d.nodes
}

func (d Doc) String() string {
	nodes := []string{}
	for _, node := range d.nodes {
		nsdfnsd := stringRecNode("", node)
		nodes = append(nodes, nsdfnsd)
	}

	return strings.Join(nodes, "\n")
}
