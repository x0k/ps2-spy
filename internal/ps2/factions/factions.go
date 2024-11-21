package ps2_factions

import "fmt"

type Id string

const (
	None Id = "0"
	VS   Id = "1"
	NC   Id = "2"
	TR   Id = "3"
	NSO  Id = "4"
)

var FactionNames = map[Id]string{
	None: "None",
	VS:   "VS",
	NC:   "NC",
	TR:   "TR",
	NSO:  "NSO",
}

func FactionNameById(id Id) string {
	if name, ok := FactionNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Faction %s", id)
}
