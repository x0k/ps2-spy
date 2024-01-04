package census2

import "io"

type queryTree struct {
	field    string
	list     Field[Bit]
	prefix   Field[Str]
	start    Field[Str]
	subTrees List[queryTree]
}

const treeFieldsSeparator = "^"
const treeKeyValueSeparator = ":"
const treeSubElementsSeparator = "'"
const treeSubTreeSeparator = ","

func Tree(field string) queryTree {
	return queryTree{
		field: field,
		list: Field[Bit]{
			name:      "list",
			separator: treeKeyValueSeparator,
		},
		prefix: Field[Str]{
			name:      "prefix",
			separator: treeKeyValueSeparator,
		},
		start: Field[Str]{
			name:      "start",
			separator: treeKeyValueSeparator,
		},
		subTrees: List[queryTree]{
			separator: treeSubTreeSeparator,
		},
	}
}

func (t queryTree) print(writer io.StringWriter) {
	writer.WriteString(t.field)
	printFields(writer, treeFieldsSeparator, treeFieldsSeparator,
		t.list,
		t.prefix,
		t.start,
	)
	if t.subTrees.isEmpty() {
		return
	}
	writer.WriteString("(")
	t.subTrees.print(writer)
	writer.WriteString(")")
}

func (t queryTree) IsList(isList bool) queryTree {
	t.list.value = Bit(isList)
	return t
}

func (t queryTree) GroupPrefix(prefix string) queryTree {
	t.prefix.value = Str(prefix)
	return t
}

func (t queryTree) StartField(field string) queryTree {
	t.start.value = Str(field)
	return t
}

func (t queryTree) WithTree(tree queryTree) queryTree {
	t.subTrees.values = append(t.subTrees.values, tree)
	return t
}
