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

type censusQuery struct {
	queryType       string
	namespace       string
	collection      string
	Terms           []CensusQueryCondition `queryProp:"conditions"`
	ExactMatchFirst bool                   `queryProp:"exactMatchFirst,default=false"`
	Timing          bool                   `queryProp:"timing,default=false"`
	IncludeNull     bool                   `queryProp:"includeNull,default=false"`
	CaseSensitive   bool                   `queryProp:"case,default=true"`
	Retry           bool                   `queryProp:"retry,default=true"`
	Limit           int                    `queryProp:"limit,default=-1"`
	LimitPerDB      int                    `queryProp:"limitPerDB,default=-1"`
	Start           int                    `queryProp:"start,default=-1"`
	Show            []string               `queryProp:"show"`
	Hide            []string               `queryProp:"hide"`
	Sort            []string               `queryProp:"sort"`
	Has             []string               `queryProp:"has"`
	Resolve         []string               `queryProp:"resolve"`
	Join            []CensusQueryJoin      `queryProp:"join"`
	Tree            []CensusQueryTree      `queryProp:"tree"`
	Distinct        string                 `queryProp:"distinct"`
	Language        string                 `queryProp:"lang"`
}

func Query(qt string, ns string, collection string) CensusQuery {
	return &censusQuery{
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

func (q *censusQuery) GetCollection() string {
	return q.collection
}

func (q *censusQuery) WithJoin(join CensusQueryJoin) CensusQuery {
	q.Join = append(q.Join, join)
	return q
}

func (q *censusQuery) WithTree(tree CensusQueryTree) CensusQuery {
	q.Tree = append(q.Tree, tree)
	return q
}

func (q *censusQuery) Where(cond CensusQueryCondition) CensusQuery {
	q.Terms = append(q.Terms, cond)
	return q
}

func (q *censusQuery) SetExactMatchFirst(exactMatchFirst bool) CensusQuery {
	q.ExactMatchFirst = exactMatchFirst
	return q
}

func (q *censusQuery) SetTiming(timing bool) CensusQuery {
	q.Timing = timing
	return q
}

func (q *censusQuery) SetIncludeNull(includeNull bool) CensusQuery {
	q.IncludeNull = includeNull
	return q
}

func (q *censusQuery) SetCase(caseSensitive bool) CensusQuery {
	q.CaseSensitive = caseSensitive
	return q
}

func (q *censusQuery) SetRetry(retry bool) CensusQuery {
	q.Retry = retry
	return q
}

func (q *censusQuery) ShowFields(fields ...string) CensusQuery {
	q.Show = append(q.Show, fields...)
	return q
}

func (q *censusQuery) HideFields(fields ...string) CensusQuery {
	q.Hide = append(q.Hide, fields...)
	return q
}

func (q *censusQuery) SortAscBy(field string) CensusQuery {
	q.Sort = append(q.Sort, field)
	return q
}

func (q *censusQuery) SortDescBy(field string) CensusQuery {
	q.Sort = append(q.Sort, field+":-1")
	return q
}

func (q *censusQuery) HasFields(fields ...string) CensusQuery {
	q.Has = append(q.Has, fields...)
	return q
}

func (q *censusQuery) SetLimit(limit int) CensusQuery {
	q.Limit = limit
	return q
}

func (q *censusQuery) SetLimitPerDB(limit int) CensusQuery {
	q.LimitPerDB = limit
	return q
}

func (q *censusQuery) SetStart(start int) CensusQuery {
	q.Start = start
	return q
}

func (q *censusQuery) AddResolve(resolves ...string) CensusQuery {
	q.Resolve = append(q.Resolve, resolves...)
	return q
}

func (q *censusQuery) SetLanguage(language string) CensusQuery {
	q.Language = language
	return q
}

func (q *censusQuery) SetDistinct(distinct string) CensusQuery {
	q.Distinct = distinct
	return q
}

func (q *censusQuery) write(builder *strings.Builder) {
	builder.WriteString(q.queryType)
	builder.WriteString("/")
	builder.WriteString(q.namespace)
	builder.WriteString("/")
	builder.WriteString(q.collection)
	writeCensusParameter(builder, q)
}

func (q *censusQuery) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
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

func (q *censusQuery) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
