package census

import (
	"reflect"
	"strings"
)

type queryTree struct {
	trees     []queryTree
	treeField string
	List      bool   `queryProp:"list"`
	Prefix    string `queryProp:"prefix"`
	Start     string `queryProp:"start"`
}

func Tree(field string) queryTree {
	return queryTree{
		treeField: field,
	}
}

func (t queryTree) IsList(isList bool) queryTree {
	return queryTree{
		trees:     t.trees,
		treeField: t.treeField,
		List:      isList,
		Prefix:    t.Prefix,
		Start:     t.Start,
	}
}

func (t queryTree) GroupPrefix(prefix string) queryTree {
	return queryTree{
		trees:     t.trees,
		treeField: t.treeField,
		List:      t.List,
		Prefix:    prefix,
		Start:     t.Start,
	}
}

func (t queryTree) StartField(field string) queryTree {
	return queryTree{
		trees:     t.trees,
		treeField: t.treeField,
		List:      t.List,
		Prefix:    t.Prefix,
		Start:     field,
	}
}

func (t queryTree) WithTree(tree queryTree) queryTree {
	return queryTree{
		trees:     append(t.trees, tree),
		treeField: t.treeField,
		List:      t.List,
		Prefix:    t.Prefix,
		Start:     t.Start,
	}
}

func (t queryTree) write(builder *strings.Builder) {
	writeCensusNestedParameter(builder, t)
}

func (t queryTree) getField() string {
	return t.treeField
}

func (t queryTree) getNestedParametersCount() int {
	return len(t.trees)
}

func (t queryTree) getNestedParameter(i int) nestedParameter {
	return t.trees[i]
}

func (t queryTree) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusParameterValue(builder, value, "'", censusValueMapperWithBitBooleans)
}
