package gokdl

import "fmt"

type Prop struct {
	Name      string
	Value     any
	valueType ValueType
}

func (p Prop) String() string {
	return fmt.Sprintf("%s=%v", p.Name, p.Value)
}

func newProp(name string, value any, t ValueType) Prop {
	return Prop{
		Name:      name,
		Value:     value,
		valueType: t,
	}
}
