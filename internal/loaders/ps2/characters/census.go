package character

import (
	"context"
	"strconv"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/ps2"
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
	Where(census2.Cond("character_id").Equals(CharacterIdOperand)).
	Resolve("outfit", "world")

func (l *CensusLoader) Load(ctx context.Context, charIds []string) (map[string]ps2.Character, error) {
	CharacterIdOperand.Set(census2.StrList(charIds...))
	chars, err := census2.ExecuteAndDecode[collections.CharacterItem](ctx, l.client, CharacterQuery)
	if err != nil {
		return nil, err
	}
	m := make(map[string]ps2.Character, len(chars))
	for _, char := range chars {
		wId, err := strconv.Atoi(char.WorldId)
		if err != nil {
			continue
		}
		m[char.CharacterId] = ps2.Character{
			Id:        char.CharacterId,
			FactionId: char.FactionId,
			Name:      char.Name.First,
			OutfitTag: char.Outfit.Alias,
			WorldId:   ps2.WorldId(wId),
		}
	}
	return m, nil
}
