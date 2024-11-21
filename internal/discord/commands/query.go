package discord_commands

type query[K comparable] struct {
	Provider string
	Key      K
}

func newQuery[K comparable](provider string, key K) query[K] {
	return query[K]{
		Provider: provider,
		Key:      key,
	}
}
