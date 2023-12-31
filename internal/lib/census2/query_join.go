package census2

import "io"

const (
	onJoinField = iota
	toJoinField
	listJoinField
	showJoinField
	hideJoinField
	injectAtJoinField
	termsJoinField
	outerJoinField
)

var queryJoinFieldNames = []string{
	"on",
	"to",
	"list",
	"show",
	"hide",
	"inject_at",
	"terms",
	"outer",
}

type queryJoin struct {
	collection    string
	fields        fields
	subJoins      extendablePrinter
	subJoinsCount int
}

func Join(collection string) queryJoin {
	return queryJoin{
		collection: collection,
		fields:     newFields(queryJoinFieldNames, "^", "^", ":", "'"),
		subJoins: List{
			separator: ",",
		},
	}
}

func (j queryJoin) print(writer io.StringWriter) {
	writer.WriteString(j.collection)
	j.fields.print(writer)
	if j.subJoinsCount > 0 {
		writer.WriteString("(")
		j.subJoins.print(writer)
		writer.WriteString(")")
	}
}
func (j queryJoin) concat(value extendablePrinter) extendablePrinter {
	j.subJoins = j.subJoins.concat(value)
	return j
}
func (j queryJoin) extend(value []extendablePrinter) extendablePrinter {
	j.subJoins = j.subJoins.extend(value)
	return j
}
func (j queryJoin) setSeparator(separator string) extendablePrinter {
	j.subJoins = j.subJoins.setSeparator(separator)
	return j
}

func (j queryJoin) IsList(isList bool) queryJoin {
	j.fields = j.fields.concatField(listJoinField, Bit(isList))
	return j
}

func (j queryJoin) IsOuter(isOuter bool) queryJoin {
	j.fields = j.fields.concatField(outerJoinField, Bit(isOuter))
	return j
}

func (j queryJoin) Show(fields ...string) queryJoin {
	j.fields = j.fields.extendListField(showJoinField, stringsToPrinters(fields))
	return j
}

func (j queryJoin) Hide(fields ...string) queryJoin {
	j.fields = j.fields.extendListField(hideJoinField, stringsToPrinters(fields))
	return j
}

func (j queryJoin) Where(term queryCondition) queryJoin {
	j.fields = j.fields.concatListField(termsJoinField, term.setSeparator("'"))
	return j
}

func (j queryJoin) On(field string) queryJoin {
	j.fields = j.fields.concatField(onJoinField, Str(field))
	return j
}

func (j queryJoin) To(field string) queryJoin {
	j.fields = j.fields.concatField(toJoinField, Str(field))
	return j
}

func (j queryJoin) InjectAt(field string) queryJoin {
	j.fields = j.fields.concatField(injectAtJoinField, Str(field))
	return j
}

func (j queryJoin) WithJoin(join queryJoin) queryJoin {
	j.subJoins = j.subJoins.concat(join)
	j.subJoinsCount++
	return j
}
