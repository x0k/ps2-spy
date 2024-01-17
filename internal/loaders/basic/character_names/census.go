package characterNames

import (
	"context"

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

var CharacterIdOperand = census2.NewPtr(census2.StrList())
var CharacterQuery = census2.NewQuery(census2.GetQuery, census2.Ps2_v2_NS, collections.Character).
	Where(census2.Cond("character_id").Equals(CharacterIdOperand)).Show("name.first")

func (l *CensusLoader) Load(ctx context.Context, charIds []string) ([]string, error) {
	CharacterIdOperand.Set(census2.StrList(charIds...))
	chars, err := census2.ExecuteAndDecode[collections.CharacterItem](ctx, l.client, CharacterQuery)
	if err != nil {
		return nil, err
	}
	names := make([]string, len(chars))
	for i, char := range chars {
		names[i] = char.Name.First
	}
	return names, nil
}
