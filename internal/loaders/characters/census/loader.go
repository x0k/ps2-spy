package census_characters_loader

import (
	"context"
	"log/slog"
	"sync"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	ps2_collections "github.com/x0k/ps2-spy/internal/lib/census2/collections/ps2"
	"github.com/x0k/ps2-spy/internal/lib/logger"
	"github.com/x0k/ps2-spy/internal/lib/retryable"
	"github.com/x0k/ps2-spy/internal/lib/retryable/perform"
	"github.com/x0k/ps2-spy/internal/lib/retryable/while"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type Loader struct {
	log                       *logger.Logger
	client                    *census2.Client
	operand                   census2.Ptr[census2.List[census2.Str]]
	queryMu                   sync.Mutex
	query                     *census2.Query
	platform                  ps2_platforms.Platform
	retryableCharactersLoader *retryable.WithArg[string, []ps2_collections.CharacterItem]
}

func New(log *logger.Logger, client *census2.Client, platform ps2_platforms.Platform) *Loader {
	operand := census2.NewPtr(census2.StrList())
	return &Loader{
		log:     log,
		client:  client,
		operand: operand,
		query: census2.NewQuery(census2.GetQuery, ps2_platforms.PlatformNamespace(platform), ps2_collections.Character).
			Where(census2.Cond("character_id").Equals(operand)).
			Show("character_id", "faction_id", "name.first").
			WithJoin(
				census2.Join(ps2_collections.OutfitMemberExtended).
					InjectAt("outfit_member_extended").
					Show("outfit_id", "alias"),
				census2.Join(ps2_collections.CharactersWorld).
					InjectAt("characters_world"),
			),
		platform: platform,
		retryableCharactersLoader: retryable.NewWithArg(
			func(ctx context.Context, url string) ([]ps2_collections.CharacterItem, error) {
				return census2.ExecutePreparedAndDecode[ps2_collections.CharacterItem](ctx, client, ps2_collections.Character, url)
			},
		),
	}
}

func (l *Loader) toUrl(charIds []census2.Str) string {
	l.queryMu.Lock()
	defer l.queryMu.Unlock()
	l.operand.Set(census2.NewList(charIds, ","))
	return l.client.ToURL(l.query)
}

func (l *Loader) load(ctx context.Context, charIds []ps2.CharacterId) ([]ps2_collections.CharacterItem, error) {
	strCharIds := make([]census2.Str, len(charIds))
	for i, charId := range charIds {
		strCharIds[i] = census2.Str(charId)
	}
	url := l.toUrl(strCharIds)
	return l.retryableCharactersLoader.Run(
		ctx,
		url,
		while.ErrorIsHere,
		while.RetryCountIsLessThan(3),
		perform.Log(
			l.log.Logger,
			slog.LevelDebug,
			"[ERROR] failed to load characters, retrying",
			slog.String("url", url),
		),
	)
}

func (l *Loader) makeCharacter(char ps2_collections.CharacterItem) ps2.Character {
	return ps2.Character{
		Id:        ps2.CharacterId(char.CharacterId),
		FactionId: ps2_factions.Id(char.FactionId),
		Name:      char.Name.First,
		OutfitId:  ps2.OutfitId(char.OutfitMemberExtended.OutfitId),
		OutfitTag: char.OutfitMemberExtended.Alias,
		WorldId:   ps2.WorldId(char.CharactersWorld.WorldId),
		Platform:  l.platform,
	}
}

func (l *Loader) Load(ctx context.Context, charIds []ps2.CharacterId) (map[ps2.CharacterId]ps2.Character, error) {
	chars, err := l.load(ctx, charIds)
	if err != nil {
		return nil, err
	}
	m := make(map[ps2.CharacterId]ps2.Character, len(charIds))
	for _, char := range chars {
		m[ps2.CharacterId(char.CharacterId)] = l.makeCharacter(char)
	}
	// If there are missing characters, then load them directly,
	// otherwise they will be skipped again.
	diff := len(charIds) - len(chars)
	if diff > 0 && len(chars) > 0 {
		missingCharIds := make([]ps2.CharacterId, 0, diff)
		for _, charId := range charIds {
			if _, ok := m[charId]; !ok {
				missingCharIds = append(missingCharIds, charId)
			}
		}
		if chars, err := l.load(ctx, missingCharIds); err == nil && len(chars) == 1 {
			for _, char := range chars {
				m[ps2.CharacterId(char.CharacterId)] = l.makeCharacter(char)
			}
		}
	}
	return m, nil
}
