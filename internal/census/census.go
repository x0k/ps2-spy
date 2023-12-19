package census

import (
	"reflect"
	"strings"
)

const VERSION = "0.0.1"

const queryTagName = "queryProp"

type censusParameter interface {
	write(builder *strings.Builder)
	writeProperty(builder *strings.Builder, key string, value reflect.Value, i int)
}

type censusQuerySearchModifier interface {
	censusParameter
	Equals(value any) censusQuerySearchModifier
	NotEquals(value any) censusQuerySearchModifier
	IsLessThan(value any) censusQuerySearchModifier
	IsLessThanOrEquals(value any) censusQuerySearchModifier
	IsGreaterThan(value any) censusQuerySearchModifier
	IsGreaterThanOrEquals(value any) censusQuerySearchModifier
	StartsWith(value any) censusQuerySearchModifier
	Contains(value any) censusQuerySearchModifier
}

type censusNestedParameter interface {
	censusParameter
	getField() string
	getNestedParametersCount() int
	getNestedParameter(i int) censusNestedParameter
}

type censusQueryTree interface {
	censusNestedParameter
	IsList(isList bool) censusQueryTree
	GroupPrefix(prefix string) censusQueryTree
	StartField(field string) censusQueryTree
	TreeField(field string) censusQueryTree
}

type censusQueryJoin interface {
	censusNestedParameter
	IsList(isList bool) censusQueryJoin
	IsOuterJoin(isOuter bool) censusQueryJoin
	ShowFields(fields ...string) censusQueryJoin
	HideFields(fields ...string) censusQueryJoin
	OnField(field string) censusQueryJoin
	ToField(field string) censusQueryJoin
	WithInjectAt(field string) censusQueryJoin
	Where(arg censusQuerySearchModifier) censusQueryJoin
	JoinCollection(collection string) censusQueryJoin
}

type censusQuery interface {
	censusParameter
	JoinCollection(join censusQueryJoin) censusQuery
	TreeField(tree censusQueryTree) censusQuery
	Where(condition censusQuerySearchModifier) censusQuery
	SetExactMatchFirst(exactMatchFirst bool) censusQuery
	SetTiming(timing bool) censusQuery
	SetIncludeNull(includeNull bool) censusQuery
	SetCase(caseSensitive bool) censusQuery
	SetRetry(retry bool) censusQuery
	ShowFields(fields ...string) censusQuery
	HideFields(fields ...string) censusQuery
	SortAscBy(sortBy string) censusQuery
	SortDescBy(sortBy string) censusQuery
	HasFields(fields ...string) censusQuery
	SetLimit(limit int) censusQuery
	SetLimitPerDB(limit int) censusQuery
	SetStart(start int) censusQuery
	AddResolve(resolves ...string) censusQuery
	SetLanguage(lang CensusLanguage) censusQuery
	SetLanguageString(lang string) censusQuery
	SetDistinct(distinct string) censusQuery
	String() string
}
