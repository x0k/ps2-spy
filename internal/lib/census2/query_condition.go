package census2

import "io"

const (
	equals                = "="
	notEquals             = "=!"
	isLessThan            = "=<"
	isLessThanOrEquals    = "=["
	isGreaterThan         = "=>"
	isGreaterThanOrEquals = "=]"
	startsWith            = "=^"
	contains              = "=*"
)

type queryCondition struct {
	field      string
	conditions List[Field[optionalPrinter]]
}

func Cond(field string) queryCondition {
	return queryCondition{
		field: field,
	}
}

func (c queryCondition) print(writer io.StringWriter) {
	c.conditions.print(writer)
}

func (c queryCondition) appendCondition(operator string, value optionalPrinter) queryCondition {
	c.conditions.values = append(c.conditions.values, Field[optionalPrinter]{
		name:      c.field,
		separator: operator,
		value:     value,
	})
	return c
}

func (c queryCondition) Equals(value optionalPrinter) queryCondition {
	return c.appendCondition(equals, value)
}

func (c queryCondition) NotEquals(value optionalPrinter) queryCondition {
	return c.appendCondition(notEquals, value)
}

func (c queryCondition) IsLessThan(value optionalPrinter) queryCondition {
	return c.appendCondition(isLessThan, value)
}

func (c queryCondition) IsLessThanOrEquals(value optionalPrinter) queryCondition {
	return c.appendCondition(isLessThanOrEquals, value)
}

func (c queryCondition) IsGreaterThan(value optionalPrinter) queryCondition {
	return c.appendCondition(isGreaterThan, value)
}

func (c queryCondition) IsGreaterThanOrEquals(value optionalPrinter) queryCondition {
	return c.appendCondition(isGreaterThanOrEquals, value)
}

func (c queryCondition) StartsWith(value optionalPrinter) queryCondition {
	return c.appendCondition(startsWith, value)
}

func (c queryCondition) Contains(value optionalPrinter) queryCondition {
	return c.appendCondition(contains, value)
}
