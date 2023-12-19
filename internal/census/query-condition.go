package census

import (
	"fmt"
	"strings"
	"time"
)

type censusOperatorType int

const (
	equals censusOperatorType = iota
	notEquals
	isLessThan
	isLessThanOrEquals
	isGreaterThan
	isGreaterThanOrEquals
	startsWith
	contains
)

var operators = []string{"", "!", "<", "[", ">", "]", "^", "*"}

type queryCondition struct {
	field    string
	value    any
	operator censusOperatorType
}

func newCensusQueryCondition(field string) censusQueryCondition {
	return &queryCondition{
		field: field,
	}
}

func (o *queryCondition) Equals(value any) {
	o.value = value
	o.operator = equals
}

func (o *queryCondition) NotEquals(value any) {
	o.value = value
	o.operator = notEquals
}

func (o *queryCondition) IsLessThan(value any) {
	o.value = value
	o.operator = isLessThan
}

func (o *queryCondition) IsLessThanOrEquals(value any) {
	o.value = value
	o.operator = isLessThanOrEquals
}

func (o *queryCondition) IsGreaterThan(value any) {
	o.value = value
	o.operator = isGreaterThan
}

func (o *queryCondition) IsGreaterThanOrEquals(value any) {
	o.value = value
	o.operator = isGreaterThanOrEquals
}

func (o *queryCondition) StartsWith(value any) {
	o.value = value
	o.operator = startsWith
}

func (o *queryCondition) Contains(value any) {
	o.value = value
	o.operator = contains
}

func (o *queryCondition) String(builder *strings.Builder) {
	builder.WriteString(o.field)
	builder.WriteString("=")
	builder.WriteString(operators[o.operator])
	builder.WriteString(o.getComparatorString())
}

func (o *queryCondition) getComparatorString() string {
	if t, ok := o.value.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("%v", o.value)
}
