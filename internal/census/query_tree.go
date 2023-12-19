package census

import (
	"reflect"
	"strings"
)

type queryTree struct {
	tree      []*queryTree
	treeField string
	List      bool   `queryProp:"list"`
	Prefix    string `queryProp:"prefix"`
	Start     string `queryProp:"start"`
}

func newCensusQueryTree(field string) censusQueryTree {
	return newQueryTree(field)
}

func newQueryTree(field string) *queryTree {
	return &queryTree{
		List:      false,
		Prefix:    "",
		Start:     "",
		tree:      make([]*queryTree, 0),
		treeField: field,
	}
}

func (t *queryTree) IsList(isList bool) censusQueryTree {
	t.List = isList
	return t
}

func (t *queryTree) GroupPrefix(prefix string) censusQueryTree {
	t.Prefix = prefix
	return t
}

func (t *queryTree) StartField(field string) censusQueryTree {
	t.Start = field
	return t
}

func (t *queryTree) TreeField(field string) censusQueryTree {
	newTree := newQueryTree(field)
	t.tree = append(t.tree, newTree)
	return newTree
}

func (t *queryTree) write(builder *strings.Builder) {
	writeCensusNestedComposableParameter(builder, t)
}

func (t *queryTree) getField() string {
	return t.treeField
}

func (t *queryTree) getNestedParametersCount() int {
	return len(t.tree)
}

func (t *queryTree) getNestedParameter(i int) censusNestedComposableParameter {
	return t.tree[i]
}

func (t *queryTree) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusComposableParameterValue(builder, value, "'")
}
