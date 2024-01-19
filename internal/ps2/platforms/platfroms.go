package platforms

import (
	"fmt"

	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
)

var ErrUnknownPlatform = fmt.Errorf("unknown platform")

const PC = "pc"
const PS4_EU = "ps4eu"
const PS4_US = "ps4us"

type PlatformItems[T any] struct {
	Pc    T
	Ps4eu T
	Ps4us T
}

var Platforms = []string{PC, PS4_EU, PS4_US}

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

var PlatformEnvironments = map[string]string{
	PC:     streaming.Ps2_env,
	PS4_EU: streaming.Ps2ps4eu_env,
	PS4_US: streaming.Ps2ps4us_env,
}

func PlatformEnvironment(platform string) (string, error) {
	if env, ok := PlatformEnvironments[platform]; ok {
		return env, nil
	}
	return "", fmt.Errorf("%s: %w", platform, ErrUnknownPlatform)
}
