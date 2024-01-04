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

type Query struct {
	queryType       string
	namespace       string
	collection      string
	terms           List[queryCondition]
	show            Field[List[Str]]
	hide            Field[List[Str]]
	sort            Field[List[Str]]
	has             Field[List[Str]]
	resolve         Field[List[Str]]
	caseSensitive   Field[BoolWithDefaultTrue]
	limit           Field[Uint]
	limitPerDB      Field[Uint]
	start           Field[Uint]
	includeNull     Field[Bool]
	language        Field[Str]
	join            Field[List[queryJoin]]
	tree            Field[List[queryTree]]
	timing          Field[Bool]
	exactMatchFirst Field[Bool]
	distinct        Field[Str]
	retry           Field[BoolWithDefaultTrue]
}

const queryKeyValueSeparator = "="
const queryFirstFieldsSeparator = "?"
const queryFieldsSeparator = "&"
const querySubElementSeparator = ","

func NewQuery(queryType, namespace, collection string) *Query {
	return &Query{
		queryType:  queryType,
		namespace:  namespace,
		collection: collection,
		terms: List[queryCondition]{
			separator: queryFieldsSeparator,
		},
		show: Field[List[Str]]{
			name:      "c:show",
			separator: queryKeyValueSeparator,
			value: List[Str]{
				separator: querySubElementSeparator,
			},
		},
		hide: Field[List[Str]]{
			name:      "c:hide",
			separator: queryKeyValueSeparator,
			value: List[Str]{
				separator: querySubElementSeparator,
			},
		},
		sort: Field[List[Str]]{
			name:      "c:sort",
			separator: queryKeyValueSeparator,
			value: List[Str]{
				separator: querySubElementSeparator,
			},
		},
		has: Field[List[Str]]{
			name:      "c:has",
			separator: queryKeyValueSeparator,
			value: List[Str]{
				separator: querySubElementSeparator,
			},
		},
		resolve: Field[List[Str]]{
			name:      "c:resolve",
			separator: queryKeyValueSeparator,
			value: List[Str]{
				separator: querySubElementSeparator,
			},
		},
		caseSensitive: Field[BoolWithDefaultTrue]{
			name:      "c:case",
			separator: queryKeyValueSeparator,
			value:     BoolWithDefaultTrue(true),
		},
		limit: Field[Uint]{
			name:      "c:limit",
			separator: queryKeyValueSeparator,
			value:     Uint(-1),
		},
		limitPerDB: Field[Uint]{
			name:      "c:limitPerDB",
			separator: queryKeyValueSeparator,
			value:     Uint(-1),
		},
		start: Field[Uint]{
			name:      "c:start",
			separator: queryKeyValueSeparator,
			value:     Uint(-1),
		},
		includeNull: Field[Bool]{
			name:      "c:includeNull",
			separator: queryKeyValueSeparator,
		},
		language: Field[Str]{
			name:      "c:lang",
			separator: queryKeyValueSeparator,
		},
		join: Field[List[queryJoin]]{
			name:      "c:join",
			separator: queryKeyValueSeparator,
			value: List[queryJoin]{
				separator: querySubElementSeparator,
			},
		},
		tree: Field[List[queryTree]]{
			name:      "c:tree",
			separator: queryKeyValueSeparator,
			value: List[queryTree]{
				separator: querySubElementSeparator,
			},
		},
		timing: Field[Bool]{
			name:      "c:timing",
			separator: queryKeyValueSeparator,
		},
		exactMatchFirst: Field[Bool]{
			name:      "c:exactMatchFirst",
			separator: queryKeyValueSeparator,
		},
		distinct: Field[Str]{
			name:      "c:distinct",
			separator: queryKeyValueSeparator,
		},
		retry: Field[BoolWithDefaultTrue]{
			name:      "c:retry",
			separator: queryKeyValueSeparator,
			value:     BoolWithDefaultTrue(true),
		},
	}
}

func NewQueryMustBeValid(queryType, namespace, collection string) *Query {
	q := NewQuery(queryType, namespace, collection)
	if err := q.Validate(); err != nil {
		panic(err)
	}
	return q
}

func (q *Query) Collection() string {
	return q.collection
}

func (q *Query) Where(term queryCondition) *Query {
	term.conditions.separator = queryFieldsSeparator
	q.terms.values = append(q.terms.values, term)
	return q
}

func (q *Query) WithJoin(join queryJoin) *Query {
	q.join.value.values = append(q.join.value.values, join)
	return q
}

func (q *Query) WithTree(tree queryTree) *Query {
	q.tree.value.values = append(q.tree.value.values, tree)
	return q
}

func (q *Query) SetExactMatchFirst(exactMatchFirst bool) *Query {
	q.exactMatchFirst.value = Bool(exactMatchFirst)
	return q
}

func (q *Query) SetTiming(timing bool) *Query {
	q.timing.value = Bool(timing)
	return q
}

func (q *Query) SetIncludeNull(includeNull bool) *Query {
	q.includeNull.value = Bool(includeNull)
	return q
}

func (q *Query) IsCaseSensitive(caseSensitive bool) *Query {
	q.caseSensitive.value = BoolWithDefaultTrue(caseSensitive)
	return q
}

func (q *Query) SetRetry(retry bool) *Query {
	q.retry.value = BoolWithDefaultTrue(retry)
	return q
}

func (q *Query) SetStart(start int) *Query {
	q.start.value = Uint(start)
	return q
}

func (q *Query) SetLimit(limit int) *Query {
	q.limit.value = Uint(limit)
	return q
}

func (q *Query) SetLimitPerDB(limit int) *Query {
	q.limitPerDB.value = Uint(limit)
	return q
}

func (q *Query) Show(fields ...string) *Query {
	q.show.value.values = append(q.show.value.values, stringsToStr(fields)...)
	return q
}

func (q *Query) Hide(fields ...string) *Query {
	q.hide.value.values = append(q.hide.value.values, stringsToStr(fields)...)
	return q
}

func (q *Query) SortAscBy(field string) *Query {
	q.sort.value.values = append(q.sort.value.values, Str(field))
	return q
}

func (q *Query) SortDescBy(field string) *Query {
	q.sort.value.values = append(q.sort.value.values, Str(field+":-1"))
	return q
}

func (q *Query) HasFields(fields ...string) *Query {
	q.has.value.values = append(q.has.value.values, stringsToStr(fields)...)
	return q
}

func (q *Query) Resolve(resolves ...string) *Query {
	q.resolve.value.values = append(q.resolve.value.values, stringsToStr(resolves)...)
	return q
}

func (q *Query) SetDistinct(distinct string) *Query {
	q.distinct.value = Str(distinct)
	return q
}

func (q *Query) SetLanguage(language string) *Query {
	q.language.value = Str(language)
	return q
}

var queryFieldNames = [...]string{
	"terms",
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

func (q *Query) fields() []optionalPrinter {
	return []optionalPrinter{
		q.terms,
		q.show,
		q.hide,
		q.sort,
		q.has,
		q.resolve,
		q.caseSensitive,
		q.limit,
		q.limitPerDB,
		q.start,
		q.includeNull,
		q.language,
		q.join,
		q.tree,
		q.timing,
		q.exactMatchFirst,
		q.distinct,
		q.retry,
	}
}

func (q *Query) print(writer io.StringWriter) {
	writer.WriteString(q.queryType)
	writer.WriteString("/")
	writer.WriteString(q.namespace)
	writer.WriteString("/")
	writer.WriteString(q.collection)
	printList(writer, queryFirstFieldsSeparator, queryFieldsSeparator, q.fields())
}

func (q *Query) String() string {
	builder := strings.Builder{}
	q.print(&builder)
	return builder.String()
}
