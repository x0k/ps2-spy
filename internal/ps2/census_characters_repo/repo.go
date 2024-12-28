package census_ps2_characters_repo

import (
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
)

type Repository struct {
	log    *logger.Logger
	client *census2.Client

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
) *Repository {
	characterIdsOperand := census2.NewPtr(census2.StrList())
	characterNamesOperand := census2.NewPtr(census2.StrList())
	return &Repository{
		log:    log,
		client: client,

		characterIdsOperand: &characterIdsOperand,
		characterIdsQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("name.first_lower").Equals(&characterIdsOperand)).Show("character_id", "name.first"),

		characterNamesOperand: &characterNamesOperand,
		characterNamesQuery: census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, ps2_collections.Character).
			Where(census2.Cond("character_id").Equals(&characterNamesOperand)).Show("character_id", "name.first"),
	}
}
