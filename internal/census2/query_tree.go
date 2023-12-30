package census2

import "io"

type queryTreeField int

const (
	listTreeField queryTreeField = iota
	prefixTreeField
	startTreeField
	treeFieldsCount
)

var treeFieldNames = [treeFieldsCount]string{
	"list",
	"prefix",
	"start",
}

type queryTree struct {
	field         string
	fields        [treeFieldsCount]printer
	fieldsCount   int
	subTrees      extendablePrinter
	subTreesCount int
}

const queryTreeFieldsSeparator = "^"
const queryTreeKeyValueSeparator = ":"
const queryTreeSubTreesSeparator = ","

func Tree(field string) queryTree {
	return queryTree{
		field: field,
		subTrees: List{
			separator: queryTreeSubTreesSeparator,
		},
	}
}

func (t queryTree) print(writer io.StringWriter) {
	writer.WriteString(t.field)
	if t.fieldsCount > 0 {
		for i := 0; i < t.fieldsCount; i++ {
			f := t.fields[i]
			if f == nil {
				continue
			}
			writer.WriteString(queryTreeFieldsSeparator)
			f.print(writer)
		}
	}
	if t.subTreesCount > 0 {
		writer.WriteString("(")
		t.subTrees.print(writer)
		writer.WriteString(")")
	}
}

func (t queryTree) setTreeField(f queryTreeField, value extendablePrinter) queryTree {
	if t.fields[f] == nil {
		t.fieldsCount++
	}
	t.fields[f] = field{
		name:      treeFieldNames[f],
		separator: queryTreeKeyValueSeparator,
		value:     value,
	}
	return t
}

func (t queryTree) IsList(isList bool) queryTree {
	return t.setTreeField(listTreeField, Bit(isList))
}

func (t queryTree) GroupPrefix(prefix string) queryTree {
	return t.setTreeField(prefixTreeField, Str(prefix))
}

func (t queryTree) StartField(field string) queryTree {
	return t.setTreeField(startTreeField, Str(field))
}

func (t queryTree) WithTree(tree queryTree) queryTree {
	t.subTreesCount++
	t.subTrees = t.subTrees.append(tree)
	return t
}
