package census_outfits_repo

import (
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
)

type Repository struct {
	client *census2.Client

	outfitIdsMu      sync.Mutex
	outfitIdsQuery   *census2.Query
	outfitIdsOperand *census2.Ptr[census2.List[census2.Str]]

	outfitTagsMu      sync.Mutex
	outfitTagsQuery   *census2.Query
	outfitTagsOperand *census2.Ptr[census2.List[census2.Str]]
}

func New(client *census2.Client) *Repository {
	outfitIdsOperand := census2.NewPtr(census2.StrList())
	outfitTagsOperand := census2.NewPtr(census2.StrList())
	return &Repository{
		client: client,

		outfitIdsOperand: &outfitIdsOperand,
		outfitIdsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("alias_lower").Equals(&outfitIdsOperand)).Show("outfit_id"),

		outfitTagsOperand: &outfitTagsOperand,
		outfitTagsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Outfit).
			Where(census2.Cond("outfit_id").Equals(&outfitTagsOperand)).Show("alias"),
	}
}
