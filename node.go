package gokdl

type Node struct {
	Name     string
	Children []Node
	Props    []Prop
	Args     []Arg
}
