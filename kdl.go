package gokdl

import (
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

func Unmarshal(bs []byte) (Document, error) {
	if !utf8.Valid(bs) {
		return Document{}, fmt.Errorf("document must contain valid UTF-8")
	}

	parser := newParser(bs)
	return parser.parse()
}

type Prop struct {
	Name  string
	Value any
}

type Arg struct {
	Value any
}

func newBoolArg(lit string) (Arg, error) {
	b, err := strconv.ParseBool(lit)
	return Arg{
		Value: b,
	}, err
}

func newArg(lit string) Arg {
	return Arg{
		Value: lit,
	}
}

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
