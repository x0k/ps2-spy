package outfit_members_saver

import (
	"github.com/x0k/ps2-spy/internal/lib/diff"
	"github.com/x0k/ps2-spy/internal/ps2"
)

const (
	OutfitMembersInitType   = "outfit_members_init"
	OutfitMembersUpdateType = "outfit_members_update"
)

type OutfitMembersInit struct {
	OutfitId ps2.OutfitId
	Members  []ps2.CharacterId
}

func (e OutfitMembersInit) Type() string {
	return OutfitMembersInitType
}

type OutfitMembersUpdate struct {
	OutfitId ps2.OutfitId
	Members  diff.Diff[ps2.CharacterId]
}

func (e OutfitMembersUpdate) Type() string {
	return OutfitMembersUpdateType
}
