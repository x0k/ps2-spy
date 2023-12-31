package census2

import "io"

const (
	listTreeField = iota
	prefixTreeField
	startTreeField
)

var queryTreeFieldNames = []string{
	"list",
	"prefix",
	"start",
}

type queryTree struct {
	field         string
	fields        fields
	subTrees      extendablePrinter
	subTreesCount int
}

func Tree(field string) queryTree {
	return queryTree{
		field:  field,
		fields: newFields(queryTreeFieldNames, "^", "^", ":", "'"),
		subTrees: List{
			separator: ",",
		},
	}
}

func (t queryTree) print(writer io.StringWriter) {
	writer.WriteString(t.field)
	t.fields.print(writer)
	if t.subTreesCount > 0 {
		writer.WriteString("(")
		t.subTrees.print(writer)
		writer.WriteString(")")
	}
}
func (t queryTree) concat(value extendablePrinter) extendablePrinter {
	t.subTrees = t.subTrees.concat(value)
	return t
}
func (t queryTree) extend(value []extendablePrinter) extendablePrinter {
	t.subTrees = t.subTrees.extend(value)
	return t
}
func (t queryTree) setSeparator(separator string) extendablePrinter {
	t.subTrees = t.subTrees.setSeparator(separator)
	return t
}

func (t queryTree) IsList(isList bool) queryTree {
	t.fields = t.fields.concatField(listTreeField, Bit(isList))
	return t
}

func (t queryTree) GroupPrefix(prefix string) queryTree {
	t.fields = t.fields.concatField(prefixTreeField, Str(prefix))
	return t
}

func (t queryTree) StartField(field string) queryTree {
	t.fields = t.fields.concatField(startTreeField, Str(field))
	return t
}

func (t queryTree) WithTree(tree queryTree) queryTree {
	t.subTrees = t.subTrees.concat(tree)
	t.subTreesCount++
	return t
}
