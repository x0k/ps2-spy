package census

import "testing"

func TestQueryLimit(t *testing.T) {
	q := newCensusQuery("test").SetLimit(100)
	s := q.String()
	e := "test?c:limit=100"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryCondition(t *testing.T) {
	q := newCensusQuery("test").Where(newCensusQueryCondition("faction_id").Equals(4))
	s := q.String()
	e := "test?faction_id=4"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}
