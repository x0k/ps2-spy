package census2

import "testing"

func TestQueryBasicParams(t *testing.T) {
	q := NewQuery(GetQuery, Ns_ps2V2, "test").
		SetExactMatchFirst(true).
		SetTiming(true).
		SetIncludeNull(true).
		IsCaseSensitive(false).
		SetRetry(false).
		SetLimit(100).
		SetLimitPerDB(20).
		SetStart(10).
		SetDistinct("foo").
		SetLanguage(LangGerman)
	s := q.String()
	e := "get/ps2:v2/test?c:exactMatchFirst=true&c:timing=true&c:includeNull=true&c:case=false&c:retry=false&c:start=10&c:limit=100&c:limitPerDB=20&c:distinct=foo&c:lang=de"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryListParams(t *testing.T) {
	q := NewQuery(GetQuery, Ns_ps2V2, "test").
		ShowFields("foo", "bar").
		HideFields("baz", "qux").
		SortAscBy("foo").
		SortDescBy("bar").
		HasFields("foo", "bar").
		Resolve("foo", "bar")
	s := q.String()
	e := "get/ps2:v2/test?c:show=foo,bar&c:hide=baz,qux&c:sort=foo,bar:-1&c:has=foo,bar&c:resolve=foo,bar"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryConditions(t *testing.T) {
	q := NewQuery(GetQuery, Ns_ps2V2, "test").
		Where(Cond("faction_id").IsLessThanOrEquals(Int(4))).
		Where(Cond("item_category_id").IsGreaterThanOrEquals(Int(2)).IsLessThan(Int(5))).
		Where(Cond("faction_id").IsGreaterThan(Int(1)))
	s := q.String()
	e := "get/ps2:v2/test?faction_id=[4&item_category_id=]2&item_category_id=<5&faction_id=>1"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}
