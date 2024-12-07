package ps2_loadout

import (
	"errors"
	"fmt"

	ps2_factions "github.com/x0k/ps2-spy/internal/ps2/factions"
)

var FactionNotFound = errors.New("faction not found")
var TypeNotFound = errors.New("type not found")

type Loadout string

type LoadoutType int

const (
	Infiltrator LoadoutType = iota
	LightAssault
	Medic
	Engineer
	HeavyAssault
	MAX
	LoadoutTypeCount
)

var toFaction = map[Loadout]ps2_factions.Id{
	"1":  ps2_factions.NC,
	"3":  ps2_factions.NC,
	"4":  ps2_factions.NC,
	"5":  ps2_factions.NC,
	"6":  ps2_factions.NC,
	"7":  ps2_factions.NC,
	"8":  ps2_factions.TR,
	"10": ps2_factions.TR,
	"11": ps2_factions.TR,
	"12": ps2_factions.TR,
	"13": ps2_factions.TR,
	"14": ps2_factions.TR,
	"15": ps2_factions.VS,
	"17": ps2_factions.VS,
	"18": ps2_factions.VS,
	"19": ps2_factions.VS,
	"20": ps2_factions.VS,
	"21": ps2_factions.VS,
	"28": ps2_factions.NSO,
	"29": ps2_factions.NSO,
	"30": ps2_factions.NSO,
	"31": ps2_factions.NSO,
	"32": ps2_factions.NSO,
	"45": ps2_factions.NSO,
}

func GetFaction(loadout Loadout) (ps2_factions.Id, error) {
	faction, ok := toFaction[loadout]
	if !ok {
		return ps2_factions.None, fmt.Errorf("%w for loadout %q", FactionNotFound, loadout)
	}
	return faction, nil
}

var toType = map[Loadout]LoadoutType{
	"1":  Infiltrator,
	"3":  LightAssault,
	"4":  Medic,
	"5":  Engineer,
	"6":  HeavyAssault,
	"7":  MAX,
	"8":  Infiltrator,
	"10": LightAssault,
	"11": Medic,
	"12": Engineer,
	"13": HeavyAssault,
	"14": MAX,
	"15": Infiltrator,
	"17": LightAssault,
	"18": Medic,
	"19": Engineer,
	"20": HeavyAssault,
	"21": MAX,
	"28": Infiltrator,
	"29": LightAssault,
	"30": Medic,
	"31": Engineer,
	"32": HeavyAssault,
	"45": MAX,
}

func GetType(loadout Loadout) (LoadoutType, error) {
	t, ok := toType[loadout]
	if !ok {
		return MAX, fmt.Errorf("%w for loadout %q", TypeNotFound, loadout)
	}
	return t, nil
}
