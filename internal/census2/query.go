package census2

import (
	"fmt"
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
	"c:distinct",
	"c:lang",
}

type field interface {
	write(builder io.StringWriter)
}

type booleanField struct {
	field        queryField
	value        bool
	defaultValue bool
}

func (b booleanField) write(builder io.StringWriter) {
	if b.value != b.defaultValue {
		builder.WriteString(fieldNames[b.field])
		builder.WriteString("=")
		if b.value {
			builder.WriteString("true")
		} else {
			builder.WriteString("false")
		}
	}
}

type intField struct {
	field queryField
	value int
}

func (i intField) write(builder io.StringWriter) {
	builder.WriteString(fieldNames[i.field])
	builder.WriteString("=")
	builder.WriteString(fmt.Sprintf("%d", i.value))
}

type stringField struct {
	field queryField
	value string
}

func (s stringField) write(builder io.StringWriter) {
	builder.WriteString(fieldNames[s.field])
	builder.WriteString("=")
	builder.WriteString(s.value)
}

type Query struct {
	queryType  string
	namespace  string
	collection string
	fields     [fieldsCount]field
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

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	q.fields[exactMatchFirstField] = booleanField{
		field: exactMatchFirstField,
		value: exactMatchFirst,
	}
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	q.fields[timingField] = booleanField{
		field: timingField,
		value: timing,
	}
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	q.fields[includeNullField] = booleanField{
		field: includeNullField,
		value: includeNull,
	}
	return q
}

func (q *Query) IsCaseSensitive(caseSensitive bool) *Query {
	q.fields[caseSensitiveField] = booleanField{
		field:        caseSensitiveField,
		value:        caseSensitive,
		defaultValue: true,
	}
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	q.fields[retryField] = booleanField{
		field:        retryField,
		value:        retry,
		defaultValue: true,
	}
	return q
}

func (q *Query) SetStart(start int) *Query {
	q.fields[startField] = intField{
		field: startField,
		value: start,
	}
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	q.fields[limitField] = intField{
		field: limitField,
		value: limit,
	}
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	q.fields[limitPerDBField] = intField{
		field: limitPerDBField,
		value: limit,
	}
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	q.fields[distinctField] = stringField{
		field: distinctField,
		value: distinct,
	}
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	q.fields[languageField] = stringField{
		field: languageField,
		value: language,
	}
	return q
}

func (q *Query) write(builder io.StringWriter) {
	builder.WriteString(q.queryType)
	builder.WriteString("/")
	builder.WriteString(q.namespace)
	builder.WriteString("/")
	builder.WriteString(q.collection)
	fields := make([]field, 0, fieldsCount)
	for i := 0; i < int(fieldsCount); i++ {
		if q.fields[i] != nil {
			fields = append(fields, q.fields[i])
		}
	}
	if len(fields) == 0 {
		return
	}
	builder.WriteString("?")
	fields[0].write(builder)
	for i := 1; i < len(fields); i++ {
		builder.WriteString("&")
		fields[i].write(builder)
	}
}

func (q *Query) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
