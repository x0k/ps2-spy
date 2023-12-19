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

type CensusQueryCondition interface {
	censusParameter
	Equals(value any) CensusQueryCondition
	NotEquals(value any) CensusQueryCondition
	IsLessThan(value any) CensusQueryCondition
	IsLessThanOrEquals(value any) CensusQueryCondition
	IsGreaterThan(value any) CensusQueryCondition
	IsGreaterThanOrEquals(value any) CensusQueryCondition
	StartsWith(value any) CensusQueryCondition
	Contains(value any) CensusQueryCondition
}

type censusNestedParameter interface {
	censusParameter
	getField() string
	getNestedParametersCount() int
	getNestedParameter(i int) censusNestedParameter
}

type CensusQueryTree interface {
	censusNestedParameter
	IsList(isList bool) CensusQueryTree
	GroupPrefix(prefix string) CensusQueryTree
	StartField(field string) CensusQueryTree
	AddTree(tree CensusQueryTree) CensusQueryTree
}

type CensusQueryJoin interface {
	censusNestedParameter
	IsList(isList bool) CensusQueryJoin
	IsOuterJoin(isOuter bool) CensusQueryJoin
	ShowFields(fields ...string) CensusQueryJoin
	HideFields(fields ...string) CensusQueryJoin
	OnField(field string) CensusQueryJoin
	ToField(field string) CensusQueryJoin
	WithInjectAt(field string) CensusQueryJoin
	Where(arg CensusQueryCondition) CensusQueryJoin
	AddJoin(join CensusQueryJoin) CensusQueryJoin
}

type CensusQuery interface {
	censusParameter
	GetCollection() string
	AddJoin(join CensusQueryJoin) CensusQuery
	AddTree(tree CensusQueryTree) CensusQuery
	Where(condition CensusQueryCondition) CensusQuery
	SetExactMatchFirst(exactMatchFirst bool) CensusQuery
	SetTiming(timing bool) CensusQuery
	SetIncludeNull(includeNull bool) CensusQuery
	SetCase(caseSensitive bool) CensusQuery
	SetRetry(retry bool) CensusQuery
	ShowFields(fields ...string) CensusQuery
	HideFields(fields ...string) CensusQuery
	SortAscBy(sortBy string) CensusQuery
	SortDescBy(sortBy string) CensusQuery
	HasFields(fields ...string) CensusQuery
	SetLimit(limit int) CensusQuery
	SetLimitPerDB(limit int) CensusQuery
	SetStart(start int) CensusQuery
	AddResolve(resolves ...string) CensusQuery
	SetLanguage(lang CensusLanguage) CensusQuery
	SetLanguageString(lang string) CensusQuery
	SetDistinct(distinct string) CensusQuery
	String() string
}

type CensusClient interface {
	Execute(query CensusQuery) (any, error)
}
