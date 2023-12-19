package census

import "testing"

func TestQueryBasicParams(t *testing.T) {
	q := NewQuery("test").
		SetExactMatchFirst(true).
		SetTiming(true).
		SetIncludeNull(true).
		SetCase(false).
		SetRetry(false).
		SetLimit(100).
		SetLimitPerDB(20).
		SetStart(10).
		SetDistinct("foo").
		SetLanguage(LangGerman)
	s := q.String()
	e := "test?c:exactMatchFirst=1&c:timing=1&c:includeNull=1&c:case=0&c:retry=0&c:limit=100&c:limitPerDB=20&c:start=10&c:distinct=foo&c:lang=de"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryConditions(t *testing.T) {
	q := NewQuery("test").
		Where(NewCond("faction_id").IsLessThan(4)).
		Where(NewCond("item_category_id").IsGreaterThan(2).IsLessThan(5)).
		Where(NewCond("faction_id").IsGreaterThan(1))
	s := q.String()
	e := "test?faction_id=<4&item_category_id=>2&item_category_id=<5&faction_id=>1"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}
