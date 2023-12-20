package census

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type censusConditionType int

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

type fieldCondition struct {
	censusParameter
	field    string
	operator string
	value    any
}

func (o *fieldCondition) write(builder *strings.Builder) {
	builder.WriteString(o.field)
	builder.WriteString(o.operator)
	builder.WriteString(o.valueAsString())
}

type queryCondition struct {
	field      string
	Conditions []*fieldCondition `queryProp:"conditions"`
}

func NewCond(field string) CensusQueryCondition {
	return &queryCondition{
		field: field,
	}
}

func (o *queryCondition) Equals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: equals,
		value:    value,
	})
	return o
}

func (o *queryCondition) NotEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: notEquals,
		value:    value,
	})
	return o
}

func (o *queryCondition) IsLessThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: isLessThan,
		value:    value,
	})
	return o
}

func (o *queryCondition) IsLessThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: isLessThanOrEquals,
		value:    value,
	})
	return o
}

func (o *queryCondition) IsGreaterThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: isGreaterThan,
		value:    value,
	})
	return o
}

func (o *queryCondition) IsGreaterThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: isGreaterThanOrEquals,
		value:    value,
	})
	return o
}

func (o *queryCondition) StartsWith(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: startsWith,
		value:    value,
	})
	return o
}

func (o *queryCondition) Contains(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:    o.field,
		operator: contains,
		value:    value,
	})
	return o
}

func (o *queryCondition) write(builder *strings.Builder) {
	writeCensusParameter(builder, o)
}

func (o *queryCondition) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	if key == "conditions" {
		writeCensusParameterValue(builder, value, "&", censusBasicValueMapper)
	}
}

func (o *fieldCondition) valueAsString() string {
	if t, ok := o.value.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("%v", o.value)
}
