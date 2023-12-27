package census

import (
	"reflect"
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

type query struct {
	queryType       string
	namespace       string
	collection      string
	Terms           []queryCondition `queryProp:"conditions"`
	ExactMatchFirst bool             `queryProp:"exactMatchFirst,default=false"`
	Timing          bool             `queryProp:"timing,default=false"`
	IncludeNull     bool             `queryProp:"includeNull,default=false"`
	CaseSensitive   bool             `queryProp:"case,default=true"`
	Retry           bool             `queryProp:"retry,default=true"`
	Limit           int              `queryProp:"limit,default=-1"`
	LimitPerDB      int              `queryProp:"limitPerDB,default=-1"`
	Start           int              `queryProp:"start,default=-1"`
	Show            []string         `queryProp:"show"`
	Hide            []string         `queryProp:"hide"`
	Sort            []string         `queryProp:"sort"`
	Has             []string         `queryProp:"has"`
	Resolve         []string         `queryProp:"resolve"`
	Join            []queryJoin      `queryProp:"join"`
	Tree            []queryTree      `queryProp:"tree"`
	Distinct        string           `queryProp:"distinct"`
	Language        string           `queryProp:"lang"`
}

func NewQuery(qt string, ns string, collection string) *query {
	return &query{
		queryType:     qt,
		namespace:     ns,
		collection:    collection,
		CaseSensitive: true,
		Retry:         true,
		Limit:         -1,
		LimitPerDB:    -1,
		Start:         -1,
	}
}

func (q *query) Collection() string {
	return q.collection
}

func (q *query) WithJoin(join queryJoin) *query {
	q.Join = append(q.Join, join)
	return q
}

func (q *query) WithTree(tree queryTree) *query {
	q.Tree = append(q.Tree, tree)
	return q
}

func (q *query) Where(cond queryCondition) *query {
	q.Terms = append(q.Terms, cond)
	return q
}

func (q *query) SetExactMatchFirst(exactMatchFirst bool) *query {
	q.ExactMatchFirst = exactMatchFirst
	return q
}

func (q *query) SetTiming(timing bool) *query {
	q.Timing = timing
	return q
}

func (q *query) SetIncludeNull(includeNull bool) *query {
	q.IncludeNull = includeNull
	return q
}

func (q *query) SetCase(caseSensitive bool) *query {
	q.CaseSensitive = caseSensitive
	return q
}

func (q *query) SetRetry(retry bool) *query {
	q.Retry = retry
	return q
}

func (q *query) ShowFields(fields ...string) *query {
	q.Show = append(q.Show, fields...)
	return q
}

func (q *query) HideFields(fields ...string) *query {
	q.Hide = append(q.Hide, fields...)
	return q
}

func (q *query) SortAscBy(field string) *query {
	q.Sort = append(q.Sort, field)
	return q
}

func (q *query) SortDescBy(field string) *query {
	q.Sort = append(q.Sort, field+":-1")
	return q
}

func (q *query) HasFields(fields ...string) *query {
	q.Has = append(q.Has, fields...)
	return q
}

func (q *query) SetLimit(limit int) *query {
	q.Limit = limit
	return q
}

func (q *query) SetLimitPerDB(limit int) *query {
	q.LimitPerDB = limit
	return q
}

func (q *query) SetStart(start int) *query {
	q.Start = start
	return q
}

func (q *query) AddResolve(resolves ...string) *query {
	q.Resolve = append(q.Resolve, resolves...)
	return q
}

func (q *query) SetLanguage(language string) *query {
	q.Language = language
	return q
}

func (q *query) SetDistinct(distinct string) *query {
	q.Distinct = distinct
	return q
}

func (q *query) write(builder *strings.Builder) {
	builder.WriteString(q.queryType)
	builder.WriteString("/")
	builder.WriteString(q.namespace)
	builder.WriteString("/")
	builder.WriteString(q.collection)
	writeCensusParameter(builder, q)
}

func (q *query) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	if i == 0 {
		builder.WriteString("?")
	} else {
		builder.WriteString("&")
	}
	if key == "conditions" {
		writeCensusParameterValue(builder, value, "&", censusBasicValueMapper)
		return
	}
	builder.WriteString("c:")
	builder.WriteString(key)
	builder.WriteString("=")
	writeCensusParameterValue(builder, value, ",", censusBasicValueMapper)
}

func (q *query) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
