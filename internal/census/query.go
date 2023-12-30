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

type Query struct {
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

func NewQuery(qt string, ns string, collection string) *Query {
	return &Query{
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

func (q *Query) Collection() string {
	return q.collection
}

func (q *Query) WithJoin(join queryJoin) *Query {
	q.Join = append(q.Join, join)
	return q
}

func (q *Query) WithTree(tree queryTree) *Query {
	q.Tree = append(q.Tree, tree)
	return q
}

func (q *Query) Where(cond queryCondition) *Query {
	q.Terms = append(q.Terms, cond)
	return q
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	q.ExactMatchFirst = exactMatchFirst
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	q.Timing = timing
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	q.IncludeNull = includeNull
	return q
}

func (q *Query) SetCase(caseSensitive bool) *Query {
	q.CaseSensitive = caseSensitive
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	q.Retry = retry
	return q
}

func (q *Query) ShowFields(fields ...string) *Query {
	q.Show = append(q.Show, fields...)
	return q
}

func (q *Query) HideFields(fields ...string) *Query {
	q.Hide = append(q.Hide, fields...)
	return q
}

func (q *Query) SortAscBy(field string) *Query {
	q.Sort = append(q.Sort, field)
	return q
}

func (q *Query) SortDescBy(field string) *Query {
	q.Sort = append(q.Sort, field+":-1")
	return q
}

func (q *Query) HasFields(fields ...string) *Query {
	q.Has = append(q.Has, fields...)
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	q.Limit = limit
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	q.LimitPerDB = limit
	return q
}

func (q *Query) SetStart(start int) *Query {
	q.Start = start
	return q
}

func (q *Query) AddResolve(resolves ...string) *Query {
	q.Resolve = append(q.Resolve, resolves...)
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	q.Language = language
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	q.Distinct = distinct
	return q
}

func (q *Query) write(builder *strings.Builder) {
	builder.WriteString(q.queryType)
	builder.WriteString("/")
	builder.WriteString(q.namespace)
	builder.WriteString("/")
	builder.WriteString(q.collection)
	writeCensusParameter(builder, q)
}

func (q *Query) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
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

func (q *Query) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
