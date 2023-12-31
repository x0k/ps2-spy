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

const (
	termsField = iota
	showQueryField
	hideQueryField
	sortQueryField
	hasQueryField
	resolveQueryField
	caseSensitiveQueryField
	limitQueryField
	limitPerDBQueryField
	startQueryField
	includeNullQueryField
	languageQueryField
	joinQueryField
	treeQueryField
	timingQueryField
	exactMatchFirstQueryField
	distinctQueryField
	retryQueryField
)

var queryFieldNames = []string{
	"__terms",
	"c:show",
	"c:hide",
	"c:sort",
	"c:has",
	"c:resolve",
	"c:case",
	"c:limit",
	"c:limitPerDB",
	"c:start",
	"c:includeNull",
	"c:lang",
	"c:join",
	"c:tree",
	"c:timing",
	"c:exactMatchFirst",
	"c:distinct",
	"c:retry",
}

type Query struct {
	queryType  string
	namespace  string
	collection string
	fields     fields
}

func NewQuery(queryType, namespace, collection string) *Query {
	return &Query{
		queryType:  queryType,
		namespace:  namespace,
		collection: collection,
		fields:     newFields(queryFieldNames, "?", "&", "=", ","),
	}
}

func (q *Query) Collection() string {
	return q.collection
}

func (q *Query) Where(term queryCondition) *Query {
	q.fields = q.fields.concatRawField(termsField, term.setSeparator("&"))
	return q
}

func (q *Query) WithJoin(join queryJoin) *Query {
	q.fields = q.fields.concatListField(joinQueryField, join)
	return q
}

func (q *Query) WithTree(tree queryTree) *Query {
	q.fields = q.fields.concatListField(treeQueryField, tree)
	return q
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	q.fields = q.fields.concatField(exactMatchFirstQueryField, Bool(exactMatchFirst))
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	q.fields = q.fields.concatField(timingQueryField, Bool(timing))
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	q.fields = q.fields.concatField(includeNullQueryField, Bool(includeNull))
	return q
}

func (q *Query) IsCaseSensitive(caseSensitive bool) *Query {
	q.fields = q.fields.concatField(caseSensitiveQueryField, Bool(caseSensitive))
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	q.fields = q.fields.concatField(retryQueryField, Bool(retry))
	return q
}

func (q *Query) SetStart(start int) *Query {
	q.fields = q.fields.concatField(startQueryField, Int(start))
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	q.fields = q.fields.concatField(limitQueryField, Int(limit))
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	q.fields = q.fields.concatField(limitPerDBQueryField, Int(limit))
	return q
}

func (q *Query) Show(fields ...string) *Query {
	q.fields = q.fields.extendListField(showQueryField, stringsToPrinters(fields))
	return q
}

func (q *Query) Hide(fields ...string) *Query {
	q.fields = q.fields.extendListField(hideQueryField, stringsToPrinters(fields))
	return q
}

func (q *Query) SortAscBy(field string) *Query {
	q.fields = q.fields.concatListField(sortQueryField, Str(field))
	return q
}

func (q *Query) SortDescBy(field string) *Query {
	q.fields = q.fields.concatListField(sortQueryField, Str(field+":-1"))
	return q
}

func (q *Query) HasFields(fields ...string) *Query {
	q.fields = q.fields.extendListField(hasQueryField, stringsToPrinters(fields))
	return q
}

func (q *Query) Resolve(resolves ...string) *Query {
	q.fields = q.fields.extendListField(resolveQueryField, stringsToPrinters(resolves))
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	q.fields = q.fields.concatField(distinctQueryField, Str(distinct))
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	q.fields = q.fields.concatField(languageQueryField, Str(language))
	return q
}

func (q *Query) print(writer io.StringWriter) {
	writer.WriteString(q.queryType)
	writer.WriteString("/")
	writer.WriteString(q.namespace)
	writer.WriteString("/")
	writer.WriteString(q.collection)
	q.fields.print(writer)
}

func (q *Query) String() string {
	builder := strings.Builder{}
	q.print(&builder)
	return builder.String()
}
