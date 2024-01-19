package factions

import "fmt"

const None = "0"
const VS = "1"
const NC = "2"
const TR = "3"
const NSO = "4"

var FactionNames = map[string]string{
	None: "None",
	VS:   "VS",
	NC:   "NC",
	TR:   "TR",
	NSO:  "NSO",
}

func FactionNameById(id string) string {
	if name, ok := FactionNames[id]; ok {
		return name
	}
	return fmt.Sprintf("Faction %s", id)
}
