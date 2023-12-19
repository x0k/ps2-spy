package census

import (
	"fmt"
	"reflect"
	"strings"
	"time"
)

type censusConditionType int

const (
	equals censusConditionType = iota
	notEquals
	isLessThan
	isLessThanOrEquals
	isGreaterThan
	isGreaterThanOrEquals
	startsWith
	contains
)

var censusConditionOperators = []string{"", "!", "<", "[", ">", "]", "^", "*"}

type fieldCondition struct {
	field        string
	operatorType censusConditionType
	value        any
}

func (o *fieldCondition) String() string {
	return fmt.Sprintf("%s=%s%v", o.field, censusConditionOperators[o.operatorType], o.valueAsString())
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
		field:        o.field,
		operatorType: equals,
		value:        value,
	})
	return o
}

func (o *queryCondition) NotEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: notEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsLessThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: isLessThan,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsLessThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: isLessThanOrEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsGreaterThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: isGreaterThan,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsGreaterThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: isGreaterThanOrEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) StartsWith(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: startsWith,
		value:        value,
	})
	return o
}

func (o *queryCondition) Contains(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		operatorType: contains,
		value:        value,
	})
	return o
}

func (o *queryCondition) write(builder *strings.Builder) {
	writeCensusParameter(builder, o)
}

func (o *queryCondition) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	writeCensusParameterValue(builder, value, "&", censusBasicValueMapper)
}

func (o *fieldCondition) valueAsString() string {
	if t, ok := o.value.(time.Time); ok {
		return t.Format("2006-01-02 15:04:05")
	}
	return fmt.Sprintf("%v", o.value)
}
