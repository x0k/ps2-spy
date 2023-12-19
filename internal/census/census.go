package census

import (
	"reflect"
	"strings"
)

const VERSION = "0.0.1"

const queryTagName = "queryProp"

type censusQueryParameter interface {
	String(builder *strings.Builder)
}

type censusComposableParameter interface {
	censusQueryParameter
	writeProperty(builder *strings.Builder, key string, value reflect.Value, i int)
}

type censusNestedComposableParameter interface {
	censusComposableParameter
	getField() string
	getSubParameters() []censusNestedComposableParameter
}

type censusQueryCondition interface {
	censusQueryParameter
	Equals(value any)
	NotEquals(value any)
	IsLessThan(value any)
	IsLessThanOrEquals(value any)
	IsGreaterThan(value any)
	IsGreaterThanOrEquals(value any)
	StartsWith(value any)
	Contains(value any)
}

type censusQueryTree interface {
	censusNestedComposableParameter
	IsList(isList bool) censusQueryTree
	GroupPrefix(prefix string) censusQueryTree
	StartField(field string) censusQueryTree
	TreeField(field string) censusQueryTree
}

type censusQueryJoin interface {
	censusNestedComposableParameter
	IsList(isList bool) censusQueryJoin
	IsOuterJoin(isOuter bool) censusQueryJoin
	ShowFields(fields ...string) censusQueryJoin
	HideFields(fields ...string) censusQueryJoin
	OnField(field string) censusQueryJoin
	ToField(field string) censusQueryJoin
	WithInjectAt(field string) censusQueryJoin
	Where(arg censusQueryParameter) censusQueryJoin
	JoinCollection(collection string) censusQueryJoin
}

type censusQuery interface {
	censusComposableParameter
	JoinCollection(join censusQueryJoin) censusQuery
	TreeField(tree censusQueryTree) censusQuery
	Where(condition censusQueryCondition) censusQuery
	ShowFields(fields ...string) censusQuery
	HideFields(fields ...string) censusQuery
	SetLimit(limit int) censusQuery
	SetStart(start int) censusQuery
	AddResolve(resolves ...string) censusQuery
	SetLanguage(lang censusLanguage) censusQuery
	SetLanguageString(lang string) censusQuery
}
