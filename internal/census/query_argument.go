package census

import "fmt"

type queryArgument struct {
	operand *queryOperand
	field   string
}

func newQueryArgument(field string) *queryArgument {
	return &queryArgument{
		field:   field,
		operand: newQueryOperand(),
	}
}

func (a *queryArgument) String() string {
	return fmt.Sprintf("%s%s", a.field, a.operand)
}
