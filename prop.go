package gokdl

import "fmt"

type Prop struct {
	Name  string
	Value any
	// ValueTypeAnnot is the type annotation for the value of the property.
	// Example: age=(u8)25
	// In this case it would be "u8".
	ValueTypeAnnot TypeAnnotation
	// TypeAnnot is the type annotation for the property itself.
	// Example: (author)name="Jonathan"
	TypeAnnot TypeAnnotation
}

func (p Prop) String() string {
	return fmt.Sprintf("%s=%v", p.Name, p.Value)
}

func newProp(
	name string,
	typeAnnot TypeAnnotation,
	value any,
	valueTypeAnnot TypeAnnotation,
) Prop {
	return Prop{
		Name:           name,
		TypeAnnot:      typeAnnot,
		Value:          value,
		ValueTypeAnnot: valueTypeAnnot,
	}
}
