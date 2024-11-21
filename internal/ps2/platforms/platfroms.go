package ps2_platforms

import (
	"github.com/x0k/ps2-spy/internal/lib/census2"
	"github.com/x0k/ps2-spy/internal/lib/census2/streaming"
)

type Platform string

const (
	PC     Platform = "pc"
	PS4_EU Platform = "ps4eu"
	PS4_US Platform = "ps4us"
)

type PlatformItems[T any] struct {
	Pc    T
	Ps4eu T
	Ps4us T
}

var Platforms = []Platform{PC, PS4_EU, PS4_US}

var PlatformNamespaces = map[Platform]string{
	PC:     census2.Ps2_v2_NS,
	PS4_EU: census2.Ps2ps4eu_v2_NS,
	PS4_US: census2.Ps2ps4us_v2_NS,
}

func PlatformNamespace(platform Platform) string {
	return PlatformNamespaces[platform]
}

var PlatformEnvironments = map[Platform]string{
	PC:     streaming.Ps2_env,
	PS4_EU: streaming.Ps2ps4eu_env,
	PS4_US: streaming.Ps2ps4us_env,
}

func PlatformEnvironment(platform Platform) string {
	return PlatformEnvironments[platform]
}
