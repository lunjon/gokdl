package gokdl

import (
	"fmt"
	"strings"
)

type Node struct {
	Name     string
	Children []Node
	Props    []Prop
	Args     []Arg
}

func stringRecNode(indentation string, n Node) string {
	double := indentation + "  "

	builder := strings.Builder{}
	args := []string{}
	for _, arg := range n.Args {
		args = append(args, fmt.Sprint(arg))
	}

	props := []string{}
	for _, prop := range n.Props {
		props = append(props, fmt.Sprintf("%s=%v", prop.Name, prop.Value))
	}

	builder.WriteString(indentation + n.Name)

	if len(args) > 0 {
		builder.WriteString("\n")
		builder.WriteString(double + "Args:  " + strings.Join(args, ", "))

	}

	if len(props) > 0 {
		builder.WriteString("\n")
		builder.WriteString(double + "Props: " + strings.Join(props, ", "))
	}

	if len(n.Children) > 0 {
		builder.WriteString("\n")
		builder.WriteString(double + "Children\n")
	}

	for _, node := range n.Children {
		sss := stringRecNode(double, node)
		builder.WriteString(sss)
		builder.WriteString("\n")
	}

	return builder.String()
}
