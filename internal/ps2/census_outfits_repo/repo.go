package ps2_census_outfits_repo

import (
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
)

type Repository struct {
	log    *logger.Logger
	client *census2.Client

	outfitIdsMu      sync.Mutex
	outfitIdsQuery   *census2.Query
	outfitIdsOperand *census2.Ptr[census2.List[census2.Str]]

	outfitTagsMu      sync.Mutex
	outfitTagsQuery   *census2.Query
	outfitTagsOperand *census2.Ptr[census2.List[census2.Str]]

	outfitMemberIdsMu      sync.Mutex
	outfitMemberIdsQuery   *census2.Query
	outfitMemberIdsOperand *census2.Ptr[census2.Str]
}

func New(
	log *logger.Logger,
	client *census2.Client,
) *Repository {
	outfitIdsOperand := census2.NewPtr(census2.StrList())
	outfitTagsOperand := census2.NewPtr(census2.StrList())
	outfitMemberIdsOperand := census2.NewPtr(census2.Str(""))
	return &Repository{
		log:    log,
		client: client,

		outfitIdsOperand: &outfitIdsOperand,
		outfitIdsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("alias_lower").Equals(&outfitIdsOperand)).Show("outfit_id", "alias"),

		outfitTagsOperand: &outfitTagsOperand,
		outfitTagsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&outfitTagsOperand)).Show("outfit_id", "alias"),

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
	}
}
