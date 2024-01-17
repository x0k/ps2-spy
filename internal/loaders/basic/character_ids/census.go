package characterIds

import (
	"context"
	"strings"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
)

type CensusLoader struct {
	client *census2.Client
}

func NewCensusLoader(client *census2.Client) *CensusLoader {
	return &CensusLoader{
		client: client,
	}
}

var CharacterNameLowerOperand = census2.NewPtr(census2.StrList())
var CharacterQuery = census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Character).
	Where(census2.Cond("name.first_lower").Equals(CharacterNameLowerOperand)).Show("character_id")

func (l *CensusLoader) Load(ctx context.Context, charNames []string) ([]string, error) {
	values := make([]census2.Str, len(charNames))
	for i, name := range charNames {
		values[i] = census2.Str(strings.ToLower(name))
	}
	CharacterNameLowerOperand.Set(census2.NewList(values, ","))
	chars, err := census2.ExecuteAndDecode[collections.CharacterItem](ctx, l.client, CharacterQuery)
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(chars))
	for i, char := range chars {
		ids[i] = char.CharacterId
	}
	return ids, nil
}
