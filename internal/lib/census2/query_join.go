package census2

import "io"

type queryJoin struct {
	collection string
	on         Field[Str]
	to         Field[Str]
	list       Field[Bit]
	show       Field[List[Str]]
	hide       Field[List[Str]]
	injectAt   Field[Str]
	terms      Field[List[queryCondition]]
	outer      Field[BitWithDefaultTrue]
	subJoins   List[queryJoin]
}

const joinKeyValueSeparator = ":"
const joinFieldsSeparator = "^"
const joinSubElementsSeparator = "'"
const joinSubJoinsSeparator = ","

func Join(collection string) queryJoin {
	return queryJoin{
		collection: collection,
		on: Field[Str]{
			name:      "on",
			separator: joinKeyValueSeparator,
		},
		to: Field[Str]{
			name:      "to",
			separator: joinKeyValueSeparator,
		},
		list: Field[Bit]{
			name:      "list",
			separator: joinKeyValueSeparator,
		},
		show: Field[List[Str]]{
			name:      "show",
			separator: joinKeyValueSeparator,
			value: List[Str]{
				separator: joinSubElementsSeparator,
			},
		},
		hide: Field[List[Str]]{
			name:      "hide",
			separator: joinKeyValueSeparator,
			value: List[Str]{
				separator: joinSubElementsSeparator,
			},
		},
		injectAt: Field[Str]{
			name:      "inject_at",
			separator: joinKeyValueSeparator,
		},
		terms: Field[List[queryCondition]]{
			name:      "terms",
			separator: joinKeyValueSeparator,
			value: List[queryCondition]{
				separator: joinSubElementsSeparator,
			},
		},
		outer: Field[BitWithDefaultTrue]{
			name:      "outer",
			separator: joinKeyValueSeparator,
			value:     BitWithDefaultTrue(true),
		},
		subJoins: List[queryJoin]{
			separator: joinSubJoinsSeparator,
		},
	}
}

func (j queryJoin) print(writer io.StringWriter) {
	writer.WriteString(j.collection)
	printList(writer, joinFieldsSeparator, joinFieldsSeparator, []optionalPrinter{
		j.on,
		j.to,
		j.list,
		j.show,
		j.hide,
		j.injectAt,
		j.terms,
		j.outer,
	})
	if j.subJoins.isEmpty() {
		return
	}
	writer.WriteString("(")
	j.subJoins.print(writer)
	writer.WriteString(")")
}

func (j queryJoin) isEmpty() bool {
	return false
}

func (j queryJoin) IsList(isList bool) queryJoin {
	j.list.value = Bit(isList)
	return j
}

func (j queryJoin) IsOuter(isOuter bool) queryJoin {
	j.outer.value = BitWithDefaultTrue(isOuter)
	return j
}

func (j queryJoin) Show(fields ...string) queryJoin {
	j.show.value.values = append(j.show.value.values, stringsToStr(fields)...)
	return j
}

func (j queryJoin) Hide(fields ...string) queryJoin {
	j.hide.value.values = append(j.hide.value.values, stringsToStr(fields)...)
	return j
}

func (j queryJoin) Where(terms ...queryCondition) queryJoin {
	for i := range terms {
		terms[i].conditions.separator = joinSubElementsSeparator
	}
	j.terms.value.values = append(j.terms.value.values, terms...)
	return j
}

func (j queryJoin) On(field string) queryJoin {
	j.on.value = Str(field)
	return j
}

func (j queryJoin) To(field string) queryJoin {
	j.to.value = Str(field)
	return j
}

func (j queryJoin) InjectAt(field string) queryJoin {
	j.injectAt.value = Str(field)
	return j
}

func (j queryJoin) WithJoin(join queryJoin) queryJoin {
	j.subJoins.values = append(j.subJoins.values, join)
	return j
}
