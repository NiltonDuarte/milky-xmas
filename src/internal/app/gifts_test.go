package app

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type giftStoreFixture struct {
	records   map[Member]core.GiftRecord
	saveError error
}

func (store *giftStoreFixture) Load(_ context.Context, member Member) (core.GiftRecord, error) {
	return store.records[member], nil
}

func (store *giftStoreFixture) Save(_ context.Context, member Member, total int64, record core.GiftRecord) (bool, error) {
	if store.saveError != nil || store.records[member].Total != total {
		return false, store.saveError
	}
	store.records[member] = record
	return true, nil
}

func TestClaimGift(t *testing.T) {
	first := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	now := first
	saveError := errors.New("disk unavailable")
	store := &giftStoreFixture{records: make(map[Member]core.GiftRecord), saveError: saveError}
	service := Service{gifts: store, now: func() time.Time { return now }}
	alice := Member{ServerID: "guild", UserID: "alice"}
	bob := Member{ServerID: "guild", UserID: "bob"}

	if claim, err := service.ClaimGift(t.Context(), alice); !errors.Is(err, saveError) || claim.Collected {
		t.Fatalf("failed save returned %+v, %v", claim, err)
	}
	store.saveError = nil
	for _, step := range []struct {
		member  Member
		elapsed time.Duration
		want    GiftClaim
	}{
		{alice, 0, GiftClaim{true, 1, first.Add(24 * time.Hour)}},
		{alice, time.Hour, GiftClaim{false, 1, first.Add(24 * time.Hour)}},
		{bob, time.Hour, GiftClaim{true, 1, first.Add(25 * time.Hour)}},
		{alice, 24 * time.Hour, GiftClaim{true, 2, first.Add(48 * time.Hour)}},
	} {
		now = first.Add(step.elapsed)
		got, err := service.ClaimGift(t.Context(), step.member)
		if err != nil || got != step.want {
			t.Fatalf("claim at %v = %+v, %v; want %+v", now, got, err, step.want)
		}
	}
}
