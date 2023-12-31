package census

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

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
	field    string
	operator string
	value    any
}

func (o fieldCondition) write(builder *strings.Builder) {
	builder.WriteString(o.field)
	builder.WriteString(o.operator)
	builder.WriteString(o.valueAsString())
}

type queryCondition struct {
	field      string
	Conditions []fieldCondition `queryProp:"conditions"`
}

func Cond(field string) queryCondition {
	return queryCondition{
		field: field,
	}
}

func (o queryCondition) Equals(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: equals,
			value:    value,
		}),
	}
}

func (o queryCondition) NotEquals(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: notEquals,
			value:    value,
		}),
	}
}

func (o queryCondition) IsLessThan(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: isLessThan,
			value:    value,
		}),
	}
}

func (o queryCondition) IsLessThanOrEquals(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: isLessThanOrEquals,
			value:    value,
		}),
	}
}

func (o queryCondition) IsGreaterThan(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: isGreaterThan,
			value:    value,
		}),
	}
}

func (o queryCondition) IsGreaterThanOrEquals(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: isGreaterThanOrEquals,
			value:    value,
		}),
	}
}

func (o queryCondition) StartsWith(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: startsWith,
			value:    value,
		}),
	}
}

func (o queryCondition) Contains(value any) queryCondition {
	return queryCondition{
		field: o.field,
		Conditions: append(o.Conditions, fieldCondition{
			field:    o.field,
			operator: contains,
			value:    value,
		}),
	}
}

func (o queryCondition) write(builder *strings.Builder) {
	writeCensusParameter(builder, o)
}

func (o queryCondition) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
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
