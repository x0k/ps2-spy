package tracking

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
)

type ErrTooManyOutfits SettingsView

func (e ErrTooManyOutfits) Error() string {
	return fmt.Sprintf("too many outfits, maximum is %d", len(e.Outfits))
}

type ErrTooManyCharacters SettingsView

func (e ErrTooManyCharacters) Error() string {
	return fmt.Sprintf("too many characters, maximum is %d", len(e.Characters))
}

type ErrFailedToIdentifyEntities struct {
	OutfitTags     []string
	FoundOutfitIds map[string]ps2.OutfitId
	CharNames      []string
	FoundCharIds   map[string]ps2.CharacterId
}

func (e ErrFailedToIdentifyEntities) Error() string {
	return "failed to identify entities"
}

type settings[C any, O any] struct {
	Characters C
	Outfits    O
}

type Settings = settings[[]ps2.CharacterId, []ps2.OutfitId]

type SettingsView = settings[[]string, []string]

type settingsDiff[C any, O any] settings[diff.Diff[C], diff.Diff[O]]

func (s settingsDiff[C, O]) IsEmpty() bool {
	return s.Characters.IsEmpty() && s.Outfits.IsEmpty()
}

type SettingsDiff = settingsDiff[ps2.CharacterId, ps2.OutfitId]

type SettingsDiffView = settingsDiff[string, string]
