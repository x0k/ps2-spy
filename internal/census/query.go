package census

import (
	"reflect"
	"strings"
)

type Query struct {
	Collection      string
	terms           []censusQueryCondition
	ExactMatchFirst bool              `queryProp:"exactMatchFirst"`
	Timing          bool              `queryProp:"timing"`
	IncludeNull     bool              `queryProp:"includeNull"`
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
	Join            []censusQueryJoin `queryProp:"join"`
	Tree            []censusQueryTree `queryProp:"tree"`
	Distinct        string            `queryProp:"distinct"`
	Language        string            `queryProp:"lang"`
}

func newCensusQuery(collection string) censusQuery {
	return &Query{
		Collection:      collection,
		terms:           make([]censusQueryCondition, 0),
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
		Join:            make([]censusQueryJoin, 0),
		Tree:            make([]censusQueryTree, 0),
		Distinct:        "",
		Language:        "",
	}
}

func (q *Query) JoinCollection(join censusQueryJoin) censusQuery {
	q.Join = append(q.Join, join)
	return q
}

func (q *Query) TreeField(tree censusQueryTree) censusQuery {
	q.Tree = append(q.Tree, tree)
	return q
}

func (q *Query) Where(cond censusQueryCondition) censusQuery {
	q.terms = append(q.terms, cond)
	return q
}

func (q *Query) ShowFields(fields ...string) censusQuery {
	q.Show = append(q.Show, fields...)
	return q
}

func (q *Query) HideFields(fields ...string) censusQuery {
	q.Hide = append(q.Hide, fields...)
	return q
}

func (q *Query) SetLimit(limit int) censusQuery {
	q.Limit = limit
	return q
}

func (q *Query) SetStart(start int) censusQuery {
	q.Start = start
	return q
}

func (q *Query) AddResolve(resolves ...string) censusQuery {
	q.Resolve = append(q.Resolve, resolves...)
	return q
}

func (q *Query) SetLanguage(language censusLanguage) censusQuery {
	q.SetLanguageString(censusLanguages[language])
	return q
}

func (q *Query) SetLanguageString(language string) censusQuery {
	q.Language = language
	return q
}

func (q *Query) String(builder *strings.Builder) {
	builder.WriteString(q.Collection)
	builder.WriteString("/")
	n := writeCensusComposableParameter(builder, q)
	if len(q.terms) == 0 {
		return
	}
	for i, t := range q.terms {
		if i == 0 && n == 0 {
			builder.WriteString("?")
		} else {
			builder.WriteString("&")
		}
		t.String(builder)
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
	writeCensusComposableParameterValue(builder, value, ",")
}
