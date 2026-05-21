package ps2_outfit_members_synchronizer

import (
	"context"
	"testing"
	"time"

	"github.com/x0k/ps2-spy/internal/lib/pubsub"
	"github.com/x0k/ps2-spy/internal/ps2"
	ps2_platforms "github.com/x0k/ps2-spy/internal/ps2/platforms"
)

type fakeOutfitsRepo struct {
	members        []ps2.CharacterId
	added          []ps2.CharacterId
	removed        []ps2.CharacterId
	synchronizedAt time.Time
	inTx           bool
}

func (r *fakeOutfitsRepo) SynchronizedAt(context.Context, ps2_platforms.Platform, ps2.OutfitId) (time.Time, error) {
	return r.synchronizedAt, nil
}

func (r *fakeOutfitsRepo) SaveSynchronizedAt(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, at time.Time) error {
	r.synchronizedAt = at
	return nil
}

func (r *fakeOutfitsRepo) TrackableOutfitIds(context.Context, ps2_platforms.Platform) ([]ps2.OutfitId, error) {
	return nil, nil
}

func (r *fakeOutfitsRepo) MemberIds(context.Context, ps2_platforms.Platform, ps2.OutfitId) ([]ps2.CharacterId, error) {
	return append([]ps2.CharacterId(nil), r.members...), nil
}

func (r *fakeOutfitsRepo) AddMember(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, member ps2.CharacterId) error {
	r.added = append(r.added, member)
	r.members = append(r.members, member)
	return nil
}

func (r *fakeOutfitsRepo) RemoveMembers(ctx context.Context, platform ps2_platforms.Platform, outfitId ps2.OutfitId, members []ps2.CharacterId) error {
	r.removed = append(r.removed, members...)
	return nil
}

func (r *fakeOutfitsRepo) Transaction(ctx context.Context, run func(r *fakeOutfitsRepo) error) error {
	r.inTx = true
	return run(r)
}

type capturePublisher struct {
	events []ps2.Event
}

func (p *capturePublisher) Publish(event pubsub.Event[ps2.EventType]) {
	p.events = append(p.events, event)
}

func TestUpdateMembersSavesTimestampWhenDiffIsEmpty(t *testing.T) {
	repo := &fakeOutfitsRepo{members: []ps2.CharacterId{"a", "b"}}
	publisher := &capturePublisher{}
	s := New(nil, repo, nil, time.Hour, publisher)
	now := time.Unix(100, 0)

	err := s.updateMembers(context.Background(), ps2_platforms.PC, "outfit", []ps2.CharacterId{"a", "b"}, now)
	if err != nil {
		t.Fatalf("updateMembers() error = %v", err)
	}
	if !repo.inTx {
		t.Fatal("updateMembers() did not run inside a transaction")
	}
	if !repo.synchronizedAt.Equal(now) {
		t.Fatalf("synchronizedAt = %v, want %v", repo.synchronizedAt, now)
	}
	if len(repo.added) != 0 || len(repo.removed) != 0 {
		t.Fatalf("unexpected member writes: added=%#v removed=%#v", repo.added, repo.removed)
	}
	if len(publisher.events) != 0 {
		t.Fatalf("unexpected events: %#v", publisher.events)
	}
}

func TestUpdateMembersPublishesDomainEventsForExistingOutfit(t *testing.T) {
	repo := &fakeOutfitsRepo{members: []ps2.CharacterId{"old", "keep"}}
	publisher := &capturePublisher{}
	s := New(nil, repo, nil, time.Hour, publisher)

	err := s.updateMembers(context.Background(), ps2_platforms.PC, "outfit", []ps2.CharacterId{"keep", "new"}, time.Unix(100, 0))
	if err != nil {
		t.Fatalf("updateMembers() error = %v", err)
	}
	if len(repo.added) != 1 || repo.added[0] != "new" {
		t.Fatalf("added = %#v", repo.added)
	}
	if len(repo.removed) != 1 || repo.removed[0] != "old" {
		t.Fatalf("removed = %#v", repo.removed)
	}
	if len(publisher.events) != 3 {
		t.Fatalf("events = %#v, want added, removed, update", publisher.events)
	}
	if _, ok := publisher.events[0].(ps2.OutfitMembersAdded); !ok {
		t.Fatalf("event[0] = %T, want OutfitMembersAdded", publisher.events[0])
	}
	if _, ok := publisher.events[1].(ps2.OutfitMembersRemoved); !ok {
		t.Fatalf("event[1] = %T, want OutfitMembersRemoved", publisher.events[1])
	}
	if _, ok := publisher.events[2].(ps2.OutfitMembersUpdate); !ok {
		t.Fatalf("event[2] = %T, want OutfitMembersUpdate", publisher.events[2])
	}
}
