package gokdl

type Node struct {
	// Name of the node.
	Name string
	// Children of the node. If the node doesn't
	// have children it is an empty list.
	Children []Node
	// Properties of the node.
	Props []Prop
	// Arguments of the node.
	Args []Arg
	// Type annotation on the node.
	// It has the zero value if no type annotation
	// exists for this node.
	TypeAnnotation TypeAnnotation
}
