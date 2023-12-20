package census

import (
	"reflect"
	"strings"
)

type queryJoin struct {
	join       []CensusQueryJoin
	collection string
	List       bool                   `queryProp:"list"`
	Outer      bool                   `queryProp:"outer,default=true"`
	Show       []string               `queryProp:"show"`
	Hide       []string               `queryProp:"hide"`
	Terms      []CensusQueryCondition `queryProp:"terms"`
	On         string                 `queryProp:"on"`
	To         string                 `queryProp:"to"`
	InjectAt   string                 `queryProp:"inject_at"`
}

func NewJoin(collection string) CensusQueryJoin {
	return &queryJoin{
		collection: collection,
		Outer:      true,
	}
}

func (j *queryJoin) IsList(isList bool) CensusQueryJoin {
	j.List = isList
	return j
}

func (j *queryJoin) IsOuterJoin(isOuter bool) CensusQueryJoin {
	j.Outer = isOuter
	return j
}

func (j *queryJoin) ShowFields(fields ...string) CensusQueryJoin {
	j.Show = fields
	return j
}

func (j *queryJoin) HideFields(fields ...string) CensusQueryJoin {
	j.Hide = fields
	return j
}

func (j *queryJoin) OnField(field string) CensusQueryJoin {
	j.On = field
	return j
}

func (j *queryJoin) ToField(field string) CensusQueryJoin {
	j.To = field
	return j
}

func (j *queryJoin) WithInjectAt(field string) CensusQueryJoin {
	j.InjectAt = field
	return j
}

func (j *queryJoin) Where(arg CensusQueryCondition) CensusQueryJoin {
	j.Terms = append(j.Terms, arg)
	return j
}

func (j *queryJoin) AddJoin(join CensusQueryJoin) CensusQueryJoin {
	j.join = append(j.join, join)
	return j
}

func (j *queryJoin) write(builder *strings.Builder) {
	writeCensusNestedParameter(builder, j)
}

func (j *queryJoin) getField() string {
	return j.collection
}

func (j *queryJoin) getNestedParametersCount() int {
	return len(j.join)
}

func (j *queryJoin) getNestedParameter(i int) censusNestedParameter {
	return j.join[i]
}

func (j *queryJoin) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusParameterValue(builder, value, "'", censusValueMapperWithBitBooleans)
}
