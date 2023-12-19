package census

import (
	"reflect"
	"strings"
)

type Query struct {
	Collection      string
	terms           []CensusQueryCondition
	ExactMatchFirst bool              `queryProp:"exactMatchFirst,default=false"`
	Timing          bool              `queryProp:"timing,default=false"`
	IncludeNull     bool              `queryProp:"includeNull,default=false"`
	CaseSensitive   bool              `queryProp:"case,default=true"`
	Retry           bool              `queryProp:"retry,default=true"`
	Limit           int               `queryProp:"limit,default=-1"`
	LimitPerDB      int               `queryProp:"limitPerDB,default=-1"`
	Start           int               `queryProp:"start,default=-1"`
	Show            []string          `queryProp:"show"`
	Hide            []string          `queryProp:"hide"`
	Sort            []string          `queryProp:"sort"`
	Has             []string          `queryProp:"has"`
	Resolve         []string          `queryProp:"resolve"`
	Join            []CensusQueryJoin `queryProp:"join"`
	Tree            []CensusQueryTree `queryProp:"tree"`
	Distinct        string            `queryProp:"distinct"`
	Language        string            `queryProp:"lang"`
}

func NewQuery(collection string) CensusQuery {
	return &Query{
		Collection:      collection,
		terms:           make([]CensusQueryCondition, 0),
		ExactMatchFirst: false,
		Timing:          false,
		IncludeNull:     false,
		CaseSensitive:   true,
		Retry:           true,
		Limit:           -1,
		LimitPerDB:      -1,
		Start:           -1,
		Show:            make([]string, 0),
		Hide:            make([]string, 0),
		Sort:            make([]string, 0),
		Has:             make([]string, 0),
		Resolve:         make([]string, 0),
		Join:            make([]CensusQueryJoin, 0),
		Tree:            make([]CensusQueryTree, 0),
		Distinct:        "",
		Language:        "",
	}
}

func (q *Query) AddJoin(join CensusQueryJoin) CensusQuery {
	q.Join = append(q.Join, join)
	return q
}

func (q *Query) AddTree(tree CensusQueryTree) CensusQuery {
	q.Tree = append(q.Tree, tree)
	return q
}

func (q *Query) Where(cond CensusQueryCondition) CensusQuery {
	q.terms = append(q.terms, cond)
	return q
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) CensusQuery {
	q.ExactMatchFirst = exactMatchFirst
	return q
}

func (q *Query) SetTiming(timing bool) CensusQuery {
	q.Timing = timing
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) CensusQuery {
	q.IncludeNull = includeNull
	return q
}

func (q *Query) SetCase(caseSensitive bool) CensusQuery {
	q.CaseSensitive = caseSensitive
	return q
}

func (q *Query) SetRetry(retry bool) CensusQuery {
	q.Retry = retry
	return q
}

func (q *Query) ShowFields(fields ...string) CensusQuery {
	q.Show = append(q.Show, fields...)
	return q
}

func (q *Query) HideFields(fields ...string) CensusQuery {
	q.Hide = append(q.Hide, fields...)
	return q
}

func (q *Query) SortAscBy(field string) CensusQuery {
	q.Sort = append(q.Sort, field)
	return q
}

func (q *Query) SortDescBy(field string) CensusQuery {
	q.Sort = append(q.Sort, field+":-1")
	return q
}

func (q *Query) HasFields(fields ...string) CensusQuery {
	q.Has = append(q.Has, fields...)
	return q
}

func (q *Query) SetLimit(limit int) CensusQuery {
	q.Limit = limit
	return q
}

func (q *Query) SetLimitPerDB(limit int) CensusQuery {
	q.LimitPerDB = limit
	return q
}

func (q *Query) SetStart(start int) CensusQuery {
	q.Start = start
	return q
}

func (q *Query) AddResolve(resolves ...string) CensusQuery {
	q.Resolve = append(q.Resolve, resolves...)
	return q
}

func (q *Query) SetLanguage(language CensusLanguage) CensusQuery {
	q.SetLanguageString(censusLanguages[language])
	return q
}

func (q *Query) SetLanguageString(language string) CensusQuery {
	q.Language = language
	return q
}

func (q *Query) SetDistinct(distinct string) CensusQuery {
	q.Distinct = distinct
	return q
}

func (q *Query) write(builder *strings.Builder) {
	builder.WriteString(q.Collection)
	n := writeCensusParameter(builder, q)
	if len(q.terms) == 0 {
		return
	}
	for i, t := range q.terms {
		if i == 0 && n == 0 {
			builder.WriteString("?")
		} else {
			builder.WriteString("&")
		}
		t.write(builder)
	}
}

func (q *Query) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	if i == 0 {
		builder.WriteString("?c:")
	} else {
		builder.WriteString("&c:")
	}
	builder.WriteString(key)
	builder.WriteString("=")
	writeCensusParameterValue(builder, value, ",", censusBasicValueMapper)
}

func (q *Query) String() string {
	builder := strings.Builder{}
	q.write(&builder)
	return builder.String()
}
