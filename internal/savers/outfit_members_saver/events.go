package outfit_members_saver

import "github.com/x0k/ps2-spy/internal/lib/diff"

const (
	OutfitMembersInitType   = "outfit_members_init"
	OutfitMembersUpdateType = "outfit_members_update"
)

type OutfitMembersInit struct {
	OutfitTag string
	Members   []string
}

func (e OutfitMembersInit) Type() string {
	return OutfitMembersInitType
}

type OutfitMembersUpdate struct {
	OutfitTag string
	Members   diff.Diff[string]
}

func (e OutfitMembersUpdate) Type() string {
	return OutfitMembersUpdateType
}
