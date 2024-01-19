package meta

import "github.com/x0k/ps2-spy/internal/lib/diff"

type TrackableEntities[T any] struct {
	Outfits    T
	Characters T
}

type SubscriptionSettings TrackableEntities[[]string]

func CalculateSubscriptionSettingsDiff(old SubscriptionSettings, new SubscriptionSettings) TrackableEntities[diff.Diff[string]] {
	return TrackableEntities[diff.Diff[string]]{
		Outfits:    diff.SlicesDiff(old.Outfits, new.Outfits),
		Characters: diff.SlicesDiff(old.Characters, new.Characters),
	}
}
