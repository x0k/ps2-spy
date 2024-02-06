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
const treeSubTreeSeparator = ","

// const treeSubElementsSeparator = "'"

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

func (t queryTree) isEmpty() bool {
	return false
}

func (t queryTree) print(writer io.StringWriter) error {
	if _, err := writer.WriteString(t.field); err != nil {
		return err
	}
	if err := printList(writer, treeFieldsSeparator, treeFieldsSeparator, []optionalPrinter{
		t.list,
		t.prefix,
		t.start,
	}); err != nil {
		return err
	}
	if t.subTrees.isEmpty() {
		return nil
	}
	if _, err := writer.WriteString("("); err != nil {
		return err
	}
	if err := t.subTrees.print(writer); err != nil {
		return err
	}
	_, err := writer.WriteString(")")
	return err
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
