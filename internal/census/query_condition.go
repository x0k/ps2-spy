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

var censusSearchModifiers = []string{"", "!", "<", "[", ">", "]", "^", "*"}

type fieldCondition struct {
	field        string
	modifierType censusConditionType
	value        any
}

func (o *fieldCondition) String() string {
	return fmt.Sprintf("%s=%s%v", o.field, censusSearchModifiers[o.modifierType], o.valueAsString())
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
		modifierType: equals,
		value:        value,
	})
	return o
}

func (o *queryCondition) NotEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: notEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsLessThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: isLessThan,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsLessThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: isLessThanOrEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsGreaterThan(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: isGreaterThan,
		value:        value,
	})
	return o
}

func (o *queryCondition) IsGreaterThanOrEquals(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: isGreaterThanOrEquals,
		value:        value,
	})
	return o
}

func (o *queryCondition) StartsWith(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: startsWith,
		value:        value,
	})
	return o
}

func (o *queryCondition) Contains(value any) CensusQueryCondition {
	o.Conditions = append(o.Conditions, &fieldCondition{
		field:        o.field,
		modifierType: contains,
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
