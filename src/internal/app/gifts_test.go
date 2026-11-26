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

func (store *giftStoreFixture) Save(_ context.Context, update GiftUpdate) (bool, error) {
	if store.saveError != nil || store.records[update.Member] != update.Previous {
		return false, store.saveError
	}
	store.records[update.Member] = update.Next
	return true, nil
}

func (store *giftStoreFixture) SaveTransfer(_ context.Context, sender, recipient GiftUpdate) (bool, error) {
	if store.saveError != nil || store.records[sender.Member] != sender.Previous || store.records[recipient.Member] != recipient.Previous {
		return false, store.saveError
	}
	store.records[sender.Member] = sender.Next
	store.records[recipient.Member] = recipient.Next
	return true, nil
}

func (store *giftStoreFixture) LastAuraSend(context.Context, Member, string, bool) (time.Time, error) {
	return time.Time{}, nil
}

func TestClaimGift(t *testing.T) {
	first := time.Date(2026, 12, 1, 18, 0, 0, 0, time.UTC)
	now := first
	roll := 9
	saveError := errors.New("disk unavailable")
	store := &giftStoreFixture{records: make(map[Member]core.GiftRecord), saveError: saveError}
	service := Service{gifts: store, tuning: core.DefaultTuning(), now: func() time.Time { return now }, random: func(n int) int {
		if n == core.GiftRarityRolls {
			return 70
		}
		if n != 10 {
			t.Fatalf("random bound = %d, want 10", n)
		}
		return roll
	}}
	alice := Member{ServerID: "guild", UserID: "alice"}
	bob := Member{ServerID: "guild", UserID: "bob"}

	if claim, err := service.ClaimGift(t.Context(), alice); !errors.Is(err, saveError) || claim != (GiftClaim{}) {
		t.Fatalf("failed save returned %+v, %v", claim, err)
	}
	store.saveError = nil
	coal := core.DailyBonus{Coal: 1} // Players at aura zero get Coal too.
	for _, step := range []struct {
		member  Member
		elapsed time.Duration
		roll    int
		want    GiftClaim
	}{
		{alice, 0, 0, GiftClaim{Collected: true, Rarity: core.RareGift, Amount: 1, Total: 1, NextClaim: first.Add(24 * time.Hour), Bonus: coal}},
		{alice, time.Hour, 5, GiftClaim{Total: 1, NextClaim: first.Add(24 * time.Hour), Remaining: 23 * time.Hour}},
		{bob, time.Hour, 2, GiftClaim{Collected: true, Rarity: core.RareGift, Amount: 3, Total: 3, NextClaim: first.Add(25 * time.Hour), Bonus: coal}},
		{alice, 24 * time.Hour, 9, GiftClaim{Collected: true, Rarity: core.RareGift, Amount: 10, Total: 11, NextClaim: first.Add(48 * time.Hour), Bonus: coal}},
	} {
		now = first.Add(step.elapsed)
		roll = step.roll
		got, err := service.ClaimGift(t.Context(), step.member)
		if err != nil || got != step.want {
			t.Fatalf("claim at %v = %+v, %v; want %+v", now, got, err, step.want)
		}
	}
}
