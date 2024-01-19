package platforms

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2"
)

var ErrUnknownPlatform = fmt.Errorf("unknown platform")

const PC = "pc"
const PS4_EU = "ps4eu"
const PS4_US = "ps4us"

var PlatformNamespaces = map[string]string{
	PC:     census2.Ps2_v2_NS,
	PS4_EU: census2.Ps2ps4eu_v2_NS,
	PS4_US: census2.Ps2ps4us_v2_NS,
}

func PlatformNamespace(platform string) (string, error) {
	if ns, ok := PlatformNamespaces[platform]; ok {
		return ns, nil
	}
	return "", fmt.Errorf("%s: %w", platform, ErrUnknownPlatform)
}
