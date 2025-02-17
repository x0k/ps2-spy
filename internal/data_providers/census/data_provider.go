package census_data_provider

import (
	"strings"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type DataProvider struct {
	log      *logger.Logger
	client   *census2.Client
	pcUrl    string
	ps4euUrl string
	ps4usUrl string

	charactersMu      sync.Mutex
	charactersQuery   *census2.Query
	charactersOperand *census2.Ptr[census2.List[census2.Str]]

	facilityMu      sync.Mutex
	facilityQuery   *census2.Query
	facilityOperand *census2.Ptr[census2.Str]

	outfitMemberIdsMu      sync.Mutex
	outfitMemberIdsQuery   *census2.Query
	outfitMemberIdsOperand *census2.Ptr[census2.Str]

	outfitsMu      sync.Mutex
	outfitsQuery   *census2.Query
	outfitsOperand *census2.Ptr[census2.List[census2.Str]]

	worldMapMu      sync.Mutex
	worldMapQuery   *census2.Query
	worldMapOperand *census2.Ptr[census2.Str]
}

func New(
	log *logger.Logger,
	client *census2.Client,
) *DataProvider {
	charactersOperand := census2.NewPtr(census2.StrList())
	facilityOperand := census2.NewPtr(census2.Str(""))
	outfitMemberIdsOperand := census2.NewPtr(census2.Str(""))
	outfitsOperand := census2.NewPtr(census2.StrList())
	worldMapOperand := census2.NewPtr(census2.Str(""))
	zoneIds := strings.Builder{}
	zoneIds.Grow(len(ps2.ZoneIds) * 3)
	zoneIds.WriteString(string(ps2.ZoneIds[0]))
	for _, zoneId := range ps2.ZoneIds[1:] {
		zoneIds.WriteByte(',')
		zoneIds.WriteString(string(zoneId))
	}
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

		charactersOperand: &charactersOperand,
		charactersQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("character_id").Equals(&charactersOperand)).
			Show("character_id", "faction_id", "name.first").
			WithJoin(
				census2.Join(ps2_collections.OutfitMemberExtended).
					InjectAt("outfit_member_extended").
					Show("outfit_id", "alias"),
				census2.Join(ps2_collections.CharactersWorld).
					InjectAt("characters_world"),
			),

		facilityOperand: &facilityOperand,
		facilityQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.MapRegion).
			Where(census2.Cond("facility_id").Equals(&facilityOperand)).
			Show("facility_id", "facility_name", "facility_type", "zone_id"),

		outfitMemberIdsOperand: &outfitMemberIdsOperand,
		outfitMemberIdsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&outfitMemberIdsOperand)).
			Show("outfit_id").
			WithJoin(
				census2.Join(ps2_collections.OutfitMember).
					Show("character_id").
					InjectAt("outfit_members").
					IsList(true),
			),

		outfitsOperand: &outfitsOperand,
		outfitsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&outfitsOperand)).
			Show("outfit_id", "name", "alias"),

		worldMapOperand: &worldMapOperand,
		worldMapQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Map).
			Where(
				census2.Cond("world_id").Equals(&worldMapOperand),
				census2.Cond("zone_ids").Equals(census2.Str(zoneIds.String())),
			).
			WithJoin(
				census2.Join(ps2_collections.MapRegion).
					Show("facility_id").
					InjectAt("map_region").
					On("Regions.Row.RowData.RegionId").
					To("map_region_id"),
			),
	}
}
