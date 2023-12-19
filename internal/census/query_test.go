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
	e := "test?c:exactMatchFirst=true&c:timing=true&c:includeNull=true&c:case=false&c:retry=false&c:limit=100&c:limitPerDB=20&c:start=10&c:distinct=foo&c:lang=de"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryListParams(t *testing.T) {
	q := NewQuery("test").
		ShowFields("foo", "bar").
		HideFields("baz", "qux").
		SortAscBy("foo").
		SortDescBy("bar").
		HasFields("foo", "bar").
		AddResolve("foo", "bar")
	s := q.String()
	e := "test?c:show=foo,bar&c:hide=baz,qux&c:sort=foo,bar:-1&c:has=foo,bar&c:resolve=foo,bar"
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

func TestQueryTree(t *testing.T) {
	// Organize a list of vehicles by type:
	q := NewQuery("vehicle").
		SetLimit(500).
		AddTree(NewTree("type_id").GroupPrefix("type_").IsList(true)).
		SetLanguage(LangEnglish)
	s := q.String()
	e := "vehicle?c:limit=500&c:tree=type_id^list:1^prefix:type_&c:lang=en"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryJoin(t *testing.T) {
	// Organize zones, map_regions, map_hexes by facility_type:
	q := NewQuery("zone").
		Where(NewCond("zone_id").Equals(2)).
		AddJoin(NewJoin("map_region").
			IsList(true).
			WithInjectAt("regions").
			HideFields("zone_id").
			AddJoin(NewJoin("map_hex").
				IsList(true).
				WithInjectAt("hex").
				HideFields("zone_id", "map_region_id"))).
		AddTree(NewTree("facility_type").
			StartField("regions").
			IsList(true)).
		SetLanguage(LangEnglish)
	s := q.String()
	e := "zone?c:join=map_region^list:1^hide:zone_id^inject_at:regions(map_hex^list:1^hide:zone_id'map_region_id^inject_at:hex)&c:tree=facility_type^list:1^start:regions&c:lang=en&zone_id=2"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}
