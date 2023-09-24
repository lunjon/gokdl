package gokdl

type Doc struct {
	nodes []Node
}

func (d Doc) Nodes() []Node {
	return d.nodes
}
