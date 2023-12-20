package census

import (
	"reflect"
	"strings"
)

type queryTree struct {
	tree      []CensusQueryTree
	treeField string
	List      bool   `queryProp:"list"`
	Prefix    string `queryProp:"prefix"`
	Start     string `queryProp:"start"`
}

func NewTree(field string) CensusQueryTree {
	return &queryTree{
		treeField: field,
	}
}

func (t *queryTree) IsList(isList bool) CensusQueryTree {
	t.List = isList
	return t
}

func (t *queryTree) GroupPrefix(prefix string) CensusQueryTree {
	t.Prefix = prefix
	return t
}

func (t *queryTree) StartField(field string) CensusQueryTree {
	t.Start = field
	return t
}

func (t *queryTree) AddTree(tree CensusQueryTree) CensusQueryTree {
	t.tree = append(t.tree, tree)
	return t
}

func (t *queryTree) write(builder *strings.Builder) {
	writeCensusNestedParameter(builder, t)
}

func (t *queryTree) getField() string {
	return t.treeField
}

func (t *queryTree) getNestedParametersCount() int {
	return len(t.tree)
}

func (t *queryTree) getNestedParameter(i int) censusNestedParameter {
	return t.tree[i]
}

func (t *queryTree) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusParameterValue(builder, value, "'", censusValueMapperWithBitBooleans)
}
