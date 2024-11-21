package census_data_provider

import (
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
)

type DataProvider struct {
	log      *logger.Logger
	client   *census2.Client
	pcUrl    string
	ps4euUrl string
	ps4usUrl string

	characterIdsMu      sync.Mutex
	characterIdsQuery   *census2.Query
	characterIdsOperand *census2.Ptr[census2.List[census2.Str]]

	characterNamesMu      sync.Mutex
	characterNamesQuery   *census2.Query
	characterNamesOperand *census2.Ptr[census2.List[census2.Str]]
}

func New(
	log *logger.Logger,
	client *census2.Client,
) *DataProvider {
	characterIdsOperand := census2.NewPtr(census2.StrList())
	characterNamesOperand := census2.NewPtr(census2.StrList())
	return &DataProvider{
		log:    log,
		client: client,
		pcUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("1,10,13,17,19,24,40"))).
			SetLimit(210)),
		ps4euUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4eu_v2_NS, ps2_collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("2000"))).
			SetLimit(30)),
		ps4usUrl: client.ToURL(census2.NewQueryMustBeValid(census2.GetQuery, census2.Ps2ps4us_v2_NS, ps2_collections.WorldEvent).
			Where(census2.Cond("type").Equals(census2.Str("METAGAME"))).
			Where(census2.Cond("world_id").Equals(census2.Str("1000"))).
			SetLimit(30)),

		characterIdsOperand: &characterIdsOperand,
		characterIdsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("name.first_lower").Equals(&characterIdsOperand)).Show("character_id"),

		characterNamesOperand: &characterNamesOperand,
		characterNamesQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("character_id").Equals(&characterNamesOperand)).Show("name.first"),
	}
}
