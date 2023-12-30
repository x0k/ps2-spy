package census2

import (
	"io"
	"strings"
)

const (
	GetQuery   = "get"
	CountQuery = "count"
)

const (
	Ns_eq2 = "eq2" //	EverQuest II	Stable version.
	// deprecated
	Ns_ps2V1      = "ps2:v1"      //	PlanetSide 2 (PC)	Deprecated. Please use ps2:v2.
	Ns_ps2V2      = "ps2:v2"      //	PlanetSide 2 (PC)	Stable version, alias is ps2.
	Ns_ps2ps4usV2 = "ps2ps4us:v2" //	US PlanetSide 2 (Playstation 4)	Stable version, alias is ps2ps4us.
	Ns_ps2ps4euV2 = "ps2ps4eu:v2" //	EU PlanetSide 2 (Playstation 4)	Stable version, alias is ps2ps4eu.
	Ns_dcuoV1     = "dcuo:v1"     //	DC Univese Online (PC and Playstation 3)	Stable version, alias dcuo.
	Ns_mtgoV1     = "mtgo:v1"     //	Magic the Gathering: Online	Stable version, alias mtgo
)

const (
	LangEnglish = "en"
	LangGerman  = "de"
	LangSpanish = "es"
	LangFrench  = "fr"
	LangItalian = "it"
	LangTurkish = "tr"
)

type queryField int

const (
	exactMatchFirstField queryField = iota
	timingField
	includeNullField
	caseSensitiveField
	retryField
	startField
	limitField
	limitPerDBField
	showField
	hideField
	sortField
	hasField
	resolveField
	distinctField
	languageField
	fieldsCount
)

var fieldNames = [fieldsCount]string{
	"c:exactMatchFirst",
	"c:timing",
	"c:includeNull",
	"c:case",
	"c:retry",
	"c:start",
	"c:limit",
	"c:limitPerDB",
	"c:show",
	"c:hide",
	"c:sort",
	"c:has",
	"c:resolve",
	"c:distinct",
	"c:lang",
}

type Query struct {
	queryType   string
	namespace   string
	collection  string
	fields      [fieldsCount]extendablePrinter
	fieldsCount int
}

func NewQuery(queryType, namespace, collection string) *Query {
	return &Query{
		queryType:  queryType,
		namespace:  namespace,
		collection: collection,
	}
}

func (q *Query) Collection() string {
	return q.collection
}

func setQueryField(q *Query, qf queryField, value extendablePrinter) {
	if q.fields[qf] == nil {
		q.fieldsCount++
	}
	q.fields[qf] = field{
		name:      fieldNames[qf],
		separator: "=",
		value:     value,
	}
}

func setExtendableQueryField(q *Query, qf queryField, printers []printer) {
	if q.fields[qf] == nil {
		q.fieldsCount++
		q.fields[qf] = field{
			name:      fieldNames[qf],
			separator: "=",
			value: list{
				values:    printers,
				separator: ",",
			},
		}
	} else {
		q.fields[qf] = q.fields[qf].extend(printers)
	}
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	setQueryField(q, exactMatchFirstField, boolField(exactMatchFirst))
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	setQueryField(q, timingField, boolField(timing))
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	setQueryField(q, includeNullField, boolField(includeNull))
	return q
}

func (q *Query) IsCaseSensitive(caseSensitive bool) *Query {
	setQueryField(q, caseSensitiveField, boolField(caseSensitive))
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	setQueryField(q, retryField, boolField(retry))
	return q
}

func (q *Query) SetStart(start int) *Query {
	setQueryField(q, startField, intField(start))
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	setQueryField(q, limitField, intField(limit))
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	setQueryField(q, limitPerDBField, intField(limit))
	return q
}

func (q *Query) ShowFields(fields ...string) *Query {
	setExtendableQueryField(q, showField, stringsToList(fields))
	return q
}

func (q *Query) HideFields(fields ...string) *Query {
	setExtendableQueryField(q, hideField, stringsToList(fields))
	return q
}

func (q *Query) SortAscBy(field string) *Query {
	setExtendableQueryField(q, sortField, []printer{printableString(field)})
	return q
}

func (q *Query) SortDescBy(field string) *Query {
	setExtendableQueryField(q, sortField, []printer{printableString(field + ":-1")})
	return q
}

func (q *Query) HasFields(fields ...string) *Query {
	setExtendableQueryField(q, hasField, stringsToList(fields))
	return q
}

func (q *Query) Resolve(resolves ...string) *Query {
	setExtendableQueryField(q, resolveField, stringsToList(resolves))
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	setQueryField(q, distinctField, printableString(distinct))
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	setQueryField(q, languageField, printableString(language))
	return q
}

func (q *Query) write(writer io.StringWriter) {
	writer.WriteString(q.queryType)
	writer.WriteString("/")
	writer.WriteString(q.namespace)
	writer.WriteString("/")
	writer.WriteString(q.collection)
	if q.fieldsCount == 0 {
		return
	}
	writer.WriteString("?")
	i := 0
	for ; i < int(fieldsCount) && q.fields[i] == nil; i++ {
	}
	q.fields[i].print(writer)
	for i++; i < int(fieldsCount); i++ {
		if q.fields[i] != nil {
			writer.WriteString("&")
			q.fields[i].print(writer)
		}
	}
}

func (q *Query) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
