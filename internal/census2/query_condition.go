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
	conditions extendablePrinter
}

func Cond(field string) queryCondition {
	return queryCondition{
		field: field,
		conditions: List{
			separator: queryFieldsSeparator,
		},
	}
}

func (c queryCondition) print(writer io.StringWriter) {
	c.conditions.print(writer)
}

func (c queryCondition) append(value printer) extendablePrinter {
	c.conditions = c.conditions.append(value)
	return c
}

func (c queryCondition) extend(value []printer) extendablePrinter {
	c.conditions = c.conditions.extend(value)
	return c
}

func (c queryCondition) appendCondition(operator string, value extendablePrinter) queryCondition {
	c.conditions = c.conditions.append(field{
		name:      c.field,
		separator: operator,
		value:     value,
	})
	return c
}

func (c queryCondition) Equals(value extendablePrinter) queryCondition {
	return c.appendCondition(equals, value)
}

func (c queryCondition) NotEquals(value extendablePrinter) queryCondition {
	return c.appendCondition(notEquals, value)
}

func (c queryCondition) IsLessThan(value extendablePrinter) queryCondition {
	return c.appendCondition(isLessThan, value)
}

func (c queryCondition) IsLessThanOrEquals(value extendablePrinter) queryCondition {
	return c.appendCondition(isLessThanOrEquals, value)
}

func (c queryCondition) IsGreaterThan(value extendablePrinter) queryCondition {
	return c.appendCondition(isGreaterThan, value)
}

func (c queryCondition) IsGreaterThanOrEquals(value extendablePrinter) queryCondition {
	return c.appendCondition(isGreaterThanOrEquals, value)
}

func (c queryCondition) StartsWith(value extendablePrinter) queryCondition {
	return c.appendCondition(startsWith, value)
}

func (c queryCondition) Contains(value extendablePrinter) queryCondition {
	return c.appendCondition(contains, value)
}
