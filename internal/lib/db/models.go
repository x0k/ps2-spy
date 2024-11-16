// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.27.0

package db

import (
	"time"
)

type ChannelToCharacter struct {
	ChannelID   string
	Platform    string
	CharacterID string
}

type ChannelToOutfit struct {
	ChannelID string
	Platform  string
	OutfitID  string
}

type Facility struct {
	FacilityID   string
	FacilityName string
	FacilityType string
	ZoneID       string
}

type Outfit struct {
	Platform   string
	OutfitID   string
	OutfitName string
	OutfitTag  string
}

type OutfitSynchronization struct {
	Platform       string
	OutfitID       string
	SynchronizedAt time.Time
}

type OutfitToCharacter struct {
	Platform    string
	OutfitID    string
	CharacterID string
}