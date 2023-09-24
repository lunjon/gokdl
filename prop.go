package gokdl

import "fmt"

type Prop struct {
	Name  string
	Value any
}

func (p Prop) String() string {
	return fmt.Sprintf("%s=%v", p.Name, p.Value)
}
