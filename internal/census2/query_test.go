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
	e := "get/ps2:v2/test?c:case=false&c:limit=100&c:limitPerDB=20&c:start=10&c:includeNull=true&c:lang=de&c:timing=true&c:exactMatchFirst=true&c:distinct=foo&c:retry=false"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryListParams(t *testing.T) {
	q := NewQuery(GetQuery, Ns_ps2V2, "test").
		Show("foo", "bar").
		Hide("baz", "qux").
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

func TestQueryTree(t *testing.T) {
	// Organize a list of vehicles by type:
	q := NewQuery(GetQuery, Ns_ps2V2, "vehicle").
		SetLimit(500).
		WithTree(Tree("type_id").GroupPrefix("type_").IsList(true)).
		SetLanguage(LangEnglish)
	s := q.String()
	e := "get/ps2:v2/vehicle?c:limit=500&c:lang=en&c:tree=type_id^list:1^prefix:type_"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryJoin(t *testing.T) {
	// Organize zones, map_regions, map_hexes by facility_type:
	q := NewQuery(GetQuery, Ns_ps2V2, "zone").
		Where(Cond("zone_id").Equals(Int(2))).
		WithJoin(Join("map_region").
			IsList(true).
			InjectAt("regions").
			Hide("zone_id").
			WithJoin(Join("map_hex").
				IsList(true).
				InjectAt("hex").
				Hide("zone_id", "map_region_id"))).
		WithTree(Tree("facility_type").
			StartField("regions").
			IsList(true)).
		SetLanguage(LangEnglish)
	s := q.String()
	e := "get/ps2:v2/zone?zone_id=2&c:lang=en&c:join=map_region^list:1^hide:zone_id^inject_at:regions(map_hex^list:1^hide:zone_id'map_region_id^inject_at:hex)&c:tree=facility_type^list:1^start:regions"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

func TestQueryInnerJoin(t *testing.T) {
	// This query looks up items unlocked by a given character but discarding any items that are not linked to a weapon
	// `IsLessThan(100)` is redundant and only for testing purposes
	q := NewQuery(GetQuery, Ns_ps2V2, "character").
		Where(Cond("name.first_lower").Equals(Str("auroram"))).
		Show("name.first", "character_id").
		WithJoin(Join("characters_item").
			IsList(true).
			InjectAt("items").
			Show("item_id").
			WithJoin(Join("item").
				Show("name.en").
				InjectAt("item_data"),
			).
			WithJoin(Join("item_to_weapon").
				On("item_id").
				To("item_id").
				Show("weapon_id").
				InjectAt("weapon").
				IsOuter(false).
				Where(Cond("weapon_id").NotEquals(Int(0)).IsLessThan(Int(100))),
			),
		)
	s := q.String()
	e := "get/ps2:v2/character?name.first_lower=auroram&c:show=name.first,character_id&c:join=characters_item^list:1^show:item_id^inject_at:items(item^show:name.en^inject_at:item_data,item_to_weapon^on:item_id^to:item_id^show:weapon_id^inject_at:weapon^terms:weapon_id=!0'weapon_id=<100^outer:0)"
	if s != e {
		t.Errorf("expected %s, got %s", e, s)
	}
}

// TODO: Omit default values
