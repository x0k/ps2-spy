package census

import (
	"reflect"
	"strings"
)

type queryJoin struct {
	joins      []queryJoin
	collection string
	List       bool             `queryProp:"list"`
	Outer      bool             `queryProp:"outer,default=true"`
	Show       []string         `queryProp:"show"`
	Hide       []string         `queryProp:"hide"`
	Terms      []queryCondition `queryProp:"terms"`
	On         string           `queryProp:"on"`
	To         string           `queryProp:"to"`
	InjectAt   string           `queryProp:"inject_at"`
}

func Join(collection string) queryJoin {
	return queryJoin{
		collection: collection,
		Outer:      true,
	}
}

func (j queryJoin) IsList(isList bool) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       isList,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) IsOuterJoin(isOuter bool) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      isOuter,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) ShowFields(fields ...string) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       fields,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) HideFields(fields ...string) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       fields,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) OnField(field string) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         field,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) ToField(field string) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         field,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) WithInjectAt(field string) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   field,
	}
}

func (j queryJoin) Where(arg queryCondition) queryJoin {
	return queryJoin{
		joins:      j.joins,
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      append(j.Terms, arg),
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) WithJoin(join queryJoin) queryJoin {
	return queryJoin{
		joins:      append(j.joins, join),
		collection: j.collection,
		List:       j.List,
		Outer:      j.Outer,
		Show:       j.Show,
		Hide:       j.Hide,
		Terms:      j.Terms,
		On:         j.On,
		To:         j.To,
		InjectAt:   j.InjectAt,
	}
}

func (j queryJoin) write(builder *strings.Builder) {
	writeCensusNestedParameter(builder, j)
}

func (j queryJoin) field() string {
	return j.collection
}

func (j queryJoin) nestedParametersCount() int {
	return len(j.joins)
}

func (j queryJoin) nestedParameter(i int) nestedParameter {
	return j.joins[i]
}

func (j queryJoin) writeProperty(builder *strings.Builder, key string, value reflect.Value, i int) {
	builder.WriteString("^")
	builder.WriteString(key)
	builder.WriteString(":")
	writeCensusParameterValue(builder, value, "'", censusValueMapperWithBitBooleans)
}
