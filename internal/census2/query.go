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
	exactMatchFirstQueryField
	timingQueryField
	includeNullQueryField
	caseSensitiveQueryField
	retryQueryField
	startQueryField
	limitQueryField
	limitPerDBQueryField
	showQueryField
	hideQueryField
	sortQueryField
	hasQueryField
	resolveQueryField
	joinQueryField
	treeQueryField
	distinctQueryField
	languageQueryField
)

var queryFieldNames = []string{
	"__terms",
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
	"c:join",
	"c:tree",
	"c:distinct",
	"c:lang",
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
	q.fields = q.fields.setRawField(termsField, term.setSeparator("&"))
	return q
}

func (q *Query) WithJoin(join queryJoin) *Query {
	q.fields = q.fields.concatField(joinQueryField, join)
	return q
}

func (q *Query) WithTree(tree queryTree) *Query {
	q.fields = q.fields.concatField(treeQueryField, tree)
	return q
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	q.fields = q.fields.setField(exactMatchFirstQueryField, Bool(exactMatchFirst))
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	q.fields = q.fields.setField(timingQueryField, Bool(timing))
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	q.fields = q.fields.setField(includeNullQueryField, Bool(includeNull))
	return q
}

func (q *Query) IsCaseSensitive(caseSensitive bool) *Query {
	q.fields = q.fields.setField(caseSensitiveQueryField, Bool(caseSensitive))
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	q.fields = q.fields.setField(retryQueryField, Bool(retry))
	return q
}

func (q *Query) SetStart(start int) *Query {
	q.fields = q.fields.setField(startQueryField, Int(start))
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	q.fields = q.fields.setField(limitQueryField, Int(limit))
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	q.fields = q.fields.setField(limitPerDBQueryField, Int(limit))
	return q
}

func (q *Query) Show(fields ...string) *Query {
	q.fields = q.fields.extendField(showQueryField, stringsToList(fields))
	return q
}

func (q *Query) Hide(fields ...string) *Query {
	q.fields = q.fields.extendField(hideQueryField, stringsToList(fields))
	return q
}

func (q *Query) SortAscBy(field string) *Query {
	q.fields = q.fields.concatField(sortQueryField, Str(field))
	return q
}

func (q *Query) SortDescBy(field string) *Query {
	q.fields = q.fields.concatField(sortQueryField, Str(field+":-1"))
	return q
}

func (q *Query) HasFields(fields ...string) *Query {
	q.fields = q.fields.extendField(hasQueryField, stringsToList(fields))
	return q
}

func (q *Query) Resolve(resolves ...string) *Query {
	q.fields = q.fields.extendField(resolveQueryField, stringsToList(resolves))
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	q.fields = q.fields.setField(distinctQueryField, Str(distinct))
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	q.fields = q.fields.setField(languageQueryField, Str(language))
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
