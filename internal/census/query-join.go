package census

import (
	"reflect"
	"strings"
)

type queryJoin struct {
	join       []censusQueryJoin
	collection string
	List       bool                   `queryProp:"list"`
	Outer      bool                   `queryProp:"outer,default=true"`
	Show       []string               `queryProp:"show"`
	Hide       []string               `queryProp:"hide"`
	Terms      []censusQueryCondition `queryProp:"terms"`
	On         string                 `queryProp:"on"`
	To         string                 `queryProp:"to"`
	InjectAt   string                 `queryProp:"inject_at"`
}

func newCensusQueryJoin(collection string) censusQueryJoin {
	return &queryJoin{
		join:       make([]censusQueryJoin, 0),
		collection: collection,
		List:       false,
		Outer:      true,
		Show:       make([]string, 0),
		Hide:       make([]string, 0),
		Terms:      make([]censusQueryCondition, 0),
		On:         "",
		To:         "",
		InjectAt:   "",
	}
}

func (j *queryJoin) IsList(isList bool) censusQueryJoin {
	j.List = isList
	return j
}

func (j *queryJoin) IsOuterJoin(isOuter bool) censusQueryJoin {
	j.Outer = isOuter
	return j
}

func (j *queryJoin) ShowFields(fields ...string) censusQueryJoin {
	j.Show = fields
	return j
}

func (j *queryJoin) HideFields(fields ...string) censusQueryJoin {
	j.Hide = fields
	return j
}

func (j *queryJoin) OnField(field string) censusQueryJoin {
	j.On = field
	return j
}

func (j *queryJoin) ToField(field string) censusQueryJoin {
	j.To = field
	return j
}

func (j *queryJoin) WithInjectAt(field string) censusQueryJoin {
	j.InjectAt = field
	return j
}

func (j *queryJoin) Where(arg censusQueryCondition) censusQueryJoin {
	j.Terms = append(j.Terms, arg)
	return j
}

func (j *queryJoin) JoinCollection(collection string) censusQueryJoin {
	newJoin := newCensusQueryJoin(collection)
	j.join = append(j.join, newJoin)
	return newJoin
}

func (j *queryJoin) String(builder *strings.Builder) {
	writeCensusNestedComposableParameter(builder, j)
}

func (j *queryJoin) getField() string {
	return j.collection
}

func (j *queryJoin) getNestedParametersCount() int {
	return len(j.join)
}

func (j *queryJoin) getNestedParameter(i int) censusNestedComposableParameter {
	return j.join[i]
}

func (j *queryJoin) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusComposableParameterValue(builder, value, "'")
}
