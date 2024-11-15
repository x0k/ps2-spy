package census2

import (
	"fmt"
	"strings"

	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
)

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

var eventTypes = map[string]struct{}{
	"battle_rank":   {},
	"battlerankup":  {},
	"battlerankups": {},

	"item":      {},
	"itemadded": {},

	"achievement":       {},
	"achievementearned": {},

	"death":  {},
	"deaths": {},

	"kill":  {},
	"kills": {},

	"vehicle_destroy": {},

	"facility_character": {},
	"playerfacility":     {},
	"facilityplayer":     {},
	"characterfacility":  {},
	"facilitycharacter":  {},
}

var groupedEventTypes = map[string]struct{}{
	"death":  {},
	"deaths": {},

	"kill":  {},
	"kills": {},
}

func (q *Query) validateCharactersThing() error {
	if q.queryType != "get" {
		return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
	}
	for i, f := range q.fields() {
		name := queryFieldNames[i]
		if !f.isEmpty() && (name == "c:join" || name == "c:tree" || name == "terms") {
			return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
		}
	}
	for _, t := range q.terms.values {
		if (t.field == "character_id" || t.field == "id") &&
			len(t.conditions.values) == 1 &&
			// TODO: Check cond value is a List
			t.conditions.values[0].separator == equalsCondition {
			continue
		}
		return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
	}
	return nil
}

func (q *Query) validateLeaderboardThing() error {
	if q.queryType != "get" {
		return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
	}
	for i, f := range q.fields() {
		name := queryFieldNames[i]
		if !f.isEmpty() && (name == "c:start" || name == "c:limit" || name == "c:join" || name == "c:tree" || name == "terms") {
			return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
		}
	}
	hasNameQuery := false
	hasPeriodQuery := false
	isChar := q.collection == ps2_collections.CharactersLeaderboard
	for _, t := range q.terms.values {
		if t.field == "world" {
			continue
		}
		if t.field == "name" && len(t.conditions.values) == 1 {
			cond := t.conditions.values[0]
			if cond.separator == equalsCondition && (cond.name == "Kills" || cond.name == "Score" || cond.name == "Time" || cond.name == "Deaths") {
				hasNameQuery = true
				continue
			}
		}
		if t.field == "period" && len(t.conditions.values) == 1 {
			cond := t.conditions.values[0]
			if cond.separator == equalsCondition && (cond.name == "Forever" || cond.name == "Monthly" || cond.name == "Weekly" || cond.name == "Daily" || cond.name == "OneLife") {
				hasPeriodQuery = true
				continue
			}
		}
		if isChar && (t.field == "character_id" || t.field == "id") &&
			len(t.conditions.values) == 1 &&
			// TODO: Check cond value is a List
			t.conditions.values[0].separator == equalsCondition {
			continue
		}
		return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
	}
	if !hasNameQuery {
		return fmt.Errorf("missing name query for collection %q", q.collection)
	}
	if !hasPeriodQuery {
		return fmt.Errorf("missing period query for collection %q", q.collection)
	}
	return nil
}

func (q *Query) validateEventThing() error {
	if q.queryType != "get" {
		return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
	}
	for i, f := range q.fields() {
		name := queryFieldNames[i]
		if !f.isEmpty() && (name == "c:join" || name == "c:tree" || name == "terms") {
			return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
		}
	}
	isChar := q.collection == ps2_collections.CharactersEvent
	isWorld := q.collection == ps2_collections.WorldEvent
	for _, t := range q.terms.values {
		if (t.field == "before" ||
			t.field == "after" ||
			t.field == "type" ||
			(isChar && (t.field == "character_id" || t.field == "id")) ||
			(isWorld && (t.field == "world_id" || t.field == "id"))) &&
			len(t.conditions.values) == 1 {
			cond := t.conditions.values[0]
			if cond.separator != equalsCondition {
				return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
			}
			if cond.name != "type" {
				continue
			}
			condStr, err := printerToString(cond.value)
			if err != nil {
				return err
			}
			types := strings.Split(condStr, ",")
			for _, tp := range types {
				if _, ok := eventTypes[strings.ToLower(tp)]; !ok {
					return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
				}
			}
		}
		return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
	}
	return nil
}

func (q *Query) Validate() error {
	// Collection:
	// map	Can only be queried by 'world_id = x' and 'zone_ids = x,y,z...'. None of the 'c:' commands are supported (except c:join, c:tree). Only 'get' is supported, 'count' is not.
	switch q.collection {
	case ps2_collections.Map:
		if q.queryType != "get" {
			return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
		}
		for i, f := range q.fields() {
			name := queryFieldNames[i]
			if !f.isEmpty() && (name == "c:join" || name == "c:tree" || name == "terms") {
				return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
			}
		}
		for _, t := range q.terms.values {
			if (t.field == "world_id" || t.field == "zone_ids") &&
				len(t.conditions.values) == 1 &&
				// TODO: Check cond value for `zone_ids` is a List
				t.conditions.values[0].separator == equalsCondition {
				continue
			}
			return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
		}
		return nil
	// characters_world	Can only be queried by 'character_id = x,y,z...' or equivalently 'id = x,y,z...'. None of the 'c:' commands are supported (except c:join, c:tree). Only 'get' is supported, 'count' is not.
	case ps2_collections.CharactersWorld:
		return q.validateCharactersThing()
	// characters_online_status	Can only be queried by 'character_id = x,y,z...' or equivalently 'id = x,y,z...'. None of the 'c:' commands are supported (except c:join, c:tree). Only 'get' is supported, 'count' is not.
	case ps2_collections.CharactersOnlineStatus:
		return q.validateCharactersThing()
	// characters_friend	Can only be queried by 'character_id = x,y,z...' or equivalently 'id = x,y,z...'. None of the 'c:' commands are supported (except c:join, c:tree). Only 'get' is supported, 'count' is not.
	case ps2_collections.CharactersFriend:
		return q.validateCharactersThing()
	// leaderboard	Can only be queried by 'name = x' (required), 'period = x' (required), 'world = [world_id]' (optional).
	// Possible values for name are: Kills, Score, Time, Deaths.
	// Possible value for period are: Forever, Monthly, Weekly, Daily, OneLife.
	// The only 'c:' commands supported are c:start and c:limit (also c:join, c:tree).
	// Only 'get' is supported, 'count' is not.
	case ps2_collections.Leaderboard:
		return q.validateLeaderboardThing()
	// characters_leaderboard	Limitations are the same as those for leaderboard
	// except 'character_id = x,y,z...' or equivalently 'id = x,y,z...' are used to limit the characters returned.
	// Please note that only the top 10,000 characters are in the leaderboard data, many characters will not have a leaderboard row.
	// Only 'get' is supported, 'count' is not.
	case ps2_collections.CharactersLeaderboard:
		return q.validateLeaderboardThing()
	// event	Can only be queried by before, after and type.
	// 'before = [timestamp]'. The before query field can be used to pull all rows by stepping through them backwards.
	// 'after = [timestamp]'. The default value of after is 0. The after query field is provided for polling purposes.
	// 'type = [BATTLE_RANK | ITEM | ACHIEVEMENT | DEATH | KILL | VEHICLE_DESTROY | FACILITY_CHARACTER]' (case-insensitive).
	//    Aliases for these types are listed below. Multiple types can be provided comma-delimited.
	//    The default value type is 'BATTLE_RANK,ACHIEVEMENT,ITEM'.
	// The only 'c:' command supported is c:limit (also c:join, c:tree). Only 'get' is supported, 'count' is not.
	case ps2_collections.Event:
		return q.validateEventThing()
	// characters_event	Limitations are the same as those for event
	// except 'character_id = x,y,z...' or equivalently 'id = x,y,z...' are used to limit the rows returned.
	case ps2_collections.CharactersEvent:
		return q.validateEventThing()
	// world_event	Limitations are the same as those for event
	// except 'world_id = x,y,z...' or equivalently 'id = x,y,z...' are used to limit the rows returned.
	case ps2_collections.WorldEvent:
		return q.validateEventThing()
	// characters_event_grouped	Can only be queried by 'character_id = x,y,z...' or equivalently 'id = x,y,z...' and 'type = [DEATH | KILL]' (case insensitive).
	//   Aliases for these types are listed below.
	//   Multiple types can be provided comma-delimited.
	//   The default value type is 'DEATH,KILL'.
	//   The only 'c:' commands supported are c:start and c:limit (also c:join, c:tree).
	//   Only 'get' is supported, 'count' is not.
	case ps2_collections.CharactersEventGrouped:
		if q.queryType != "get" {
			return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
		}
		for i, f := range q.fields() {
			name := queryFieldNames[i]
			if !f.isEmpty() && (name == "c:start" || name == "c:limit" || name == "c:join" || name == "c:tree" || name == "terms") {
				return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
			}
		}
		for _, t := range q.terms.values {
			if (t.field == "character_id" || t.field == "id" || t.field == "type") && len(t.conditions.values) == 1 {
				cond := t.conditions.values[0]
				if cond.separator != equalsCondition {
					return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
				}
				if cond.name != "type" {
					continue
				}
				condStr, err := printerToString(cond.value)
				if err != nil {
					return err
				}
				types := strings.Split(condStr, ",")
				for _, tp := range types {
					if _, ok := groupedEventTypes[strings.ToLower(tp)]; !ok {
						return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
					}
				}
			}
			return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
		}
		return nil
	// single_character_by_id	Can only be queried by 'character_id = x' or equivalently 'id = x'.
	// None of the 'c:' commands are supported (except c:join, c:tree).
	// Only 'get' is supported, 'count' is not.
	case ps2_collections.SingleCharacterById:
		if q.queryType != "get" {
			return fmt.Errorf("invalid query type %q for collection %q", q.queryType, q.collection)
		}
		for i, f := range q.fields() {
			name := queryFieldNames[i]
			if !f.isEmpty() && (name == "c:join" || name == "c:tree" || name == "terms") {
				return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
			}
		}
		for _, t := range q.terms.values {
			if (t.field == "character_id" || t.field == "id") && len(t.conditions.values) == 1 && t.conditions.values[0].separator == equalsCondition {
				continue
			}
			return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
		}
		return nil
	// characters_item	Can only be queried by 'character_id = x,y,z...' or equivalently 'id = x,y,z...'.
	// None of the 'c:' commands are supported (except c:join, c:tree).
	case ps2_collections.CharactersItem:
		for i, f := range q.fields() {
			name := queryFieldNames[i]
			if !f.isEmpty() && (name == "c:join" || name == "c:tree" || name == "terms") {
				return fmt.Errorf("invalid field %q for collection %q", queryFieldNames[i], q.collection)
			}
		}
		for _, t := range q.terms.values {
			if (t.field == "character_id" || t.field == "id") && len(t.conditions.values) == 1 && t.conditions.values[0].separator == equalsCondition {
				continue
			}
			return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
		}
		return nil
	// world Querying by name.en, name.fr, etc is not supported.
	case ps2_collections.World:
		for _, t := range q.terms.values {
			if strings.HasPrefix(t.field, "name.") {
				return fmt.Errorf("invalid field %q for collection %q", t.field, q.collection)
			}
		}
		return nil
	}
	return nil
}
