package census2

// import (
// 	"context"
// 	"net/http"
// 	"testing"
// )

// func TestWorldsPopulation(t *testing.T) {
// 	c := NewClient("https://census.daybreakgames.com", "", &http.Client{})
// 	q := NewQuery(GetQuery, Ns_ps2V2, WorldEventCollection).
// 		Where(Cond("type").Equals(Str("METAGAME"))).
// 		Where(Cond("world_id").Equals(Str("1,10,13,17,19,24,40,1000,2000"))).
// 		SetLimit(100)
// 	state, err := ExecuteAndDecode[WorldEvent](context.Background(), c, q)
// 	if err != nil {
// 		t.Fatal(err)
// 	}
// 	t.Log(state)
// }
