package characters_loader

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
	"github.com/x0k/ps2-spy/internal/ps2/factions"
	"github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type CensusLoader struct {
	log                       *slog.Logger
	client                    *census2.Client
	operand                   census2.Ptr[census2.List[census2.Str]]
	queryMu                   sync.Mutex
	query                     *census2.Query
	platform                  platforms.Platform
	retryableCharactersLoader *retryable.WithArg[string, []collections.CharacterItem]
}

func NewCensus(log *slog.Logger, client *census2.Client, platform platforms.Platform) *CensusLoader {
	operand := census2.NewPtr(census2.StrList())
	return &CensusLoader{
		log: log.With(
			slog.String("component", "loaders.characters_loader.CensusLoader"),
			slog.String("platform", string(platform)),
		),
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, platforms.PlatformNamespace(platform), collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).
			Show("character_id", "faction_id", "name.first").
			WithJoin(
				census2.Join(collections.OutfitMemberExtended).
					InjectAt("outfit_member_extended").
					Show("outfit_id", "alias"),
				census2.Join(collections.CharactersWorld).
					InjectAt("characters_world"),
			),
		platform: platform,
		retryableCharactersLoader: retryable.NewWithArg(
			func(ctx context.Context, url string) ([]collections.CharacterItem, error) {
				return census2.ExecutePreparedAndDecode[collections.CharacterItem](ctx, client, collections.Character, url)
			},
		),
	}
}

func (l *CensusLoader) toUrl(charIds []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.query)
}

func (l *CensusLoader) Load(ctx context.Context, charIds []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
	strCharIds := make([]census2.Str, len(charIds))
	for i, charId := range charIds {
		strCharIds[i] = census2.Str(charId)
	}
	url := l.toUrl(strCharIds)
	chars, err := l.retryableCharactersLoader.Run(
		ctx,
		url,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Log(
			l.log,
			slog.LevelDebug,
			"[ERROR] failed to load characters, retrying",
			slog.String("url", url),
		),
	)
	if err != nil {
		return nil, err
	}
	m := make(map[ps2.CharacterId]ps2.Character, len(chars))
	for _, char := range chars {
		cId := ps2.CharacterId(char.CharacterId)
		m[cId] = ps2.Character{
			Id:        cId,
			FactionId: factions.Id(char.FactionId),
			Name:      char.Name.First,
			OutfitId:  ps2.OutfitId(char.OutfitMemberExtended.OutfitId),
			OutfitTag: char.OutfitMemberExtended.Alias,
			WorldId:   ps2.WorldId(char.CharactersWorld.WorldId),
			Platform:  l.platform,
		}
	}
	return m, nil
}
