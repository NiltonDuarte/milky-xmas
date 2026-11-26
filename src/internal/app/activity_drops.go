package app

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

var (
	ErrDropUnavailable = errors.New("gift drop is no longer available")
	ErrDropCollected   = errors.New("gift drop already collected")
	ErrDropCooldown    = errors.New("gift rain collection is cooling down")
	ErrDropSlotTaken   = errors.New("gift drop slot already opened")
)

type GiftDrop struct {
	ID   string
	Kind core.GiftDropKind
}

type dropChannel struct{ server, channel string }

type dropActivity struct {
	messages int
	target   int
}

type liveDrop struct {
	GiftDrop
	createdAt time.Time
	claims    map[string]time.Time
	taken     uint8 // Opened slots of a Random Active Drop, one bit per slot.
	results   []DropClaimRecord
}

// DropClaimRecord is one collected prize, in collection order.
type DropClaimRecord struct {
	UserID string
	Coal   bool
	Rarity core.GiftRarity
}

// DropSummary lists who collected what from a finished drop.
type DropSummary struct {
	Kind   core.GiftDropKind
	Claims []DropClaimRecord
}

// ActivityDrops owns short-lived channel activity and per-player claim state.
// Inventory awards remain durable in the Service's GiftStore.
type ActivityDrops struct {
	service    Service
	mu         sync.Mutex
	sequence   uint64
	activity   map[dropChannel]dropActivity
	drops      map[string]liveDrop
	rainClaims map[Member]time.Time
}

// Lifetime is how long a drop stays open.
func (drops *ActivityDrops) Lifetime() time.Duration {
	return drops.service.tuning.Drops.Lifetime
}

func NewActivityDrops(service Service) *ActivityDrops {
	return &ActivityDrops{
		service:    service,
		activity:   make(map[dropChannel]dropActivity),
		drops:      make(map[string]liveDrop),
		rainClaims: make(map[Member]time.Time),
	}
}

// Observe records one eligible human message and returns at most one new drop.
func (drops *ActivityDrops) Observe(server, channel string) *GiftDrop {
	if drops == nil || server == "" || channel == "" {
		return nil
	}
	drops.mu.Lock()
	defer drops.mu.Unlock()
	drops.pruneLocked(drops.service.now().UTC())

	key := dropChannel{server: server, channel: channel}
	activity := drops.activity[key]
	tuning := drops.service.tuning.Drops
	if activity.target == 0 {
		activity.target = tuning.NextGiftRainThreshold(drops.service.random(tuning.GiftRainMessageRange()))
	}
	activity.messages++
	if activity.messages >= activity.target {
		activity.messages = 0
		activity.target = tuning.NextGiftRainThreshold(drops.service.random(tuning.GiftRainMessageRange()))
		drops.activity[key] = activity
		return drops.createLocked(core.GiftRainDrop)
	}
	drops.activity[key] = activity
	if tuning.RandomActiveDropEligible(drops.service.random(tuning.RandomActiveOneIn)) {
		return drops.createLocked(core.RandomActiveDrop)
	}
	return nil
}

func (drops *ActivityDrops) createLocked(kind core.GiftDropKind) *GiftDrop {
	drops.sequence++
	drop := GiftDrop{ID: strconv.FormatUint(drops.sequence, 10), Kind: kind}
	drops.drops[drop.ID] = liveDrop{GiftDrop: drop, createdAt: drops.service.now().UTC(), claims: make(map[string]time.Time)}
	return &drop
}

func (drops *ActivityDrops) pruneLocked(now time.Time) {
	for id, drop := range drops.drops {
		// Adapters close drops themselves at expiry. This only drops forgotten ones.
		if now.Sub(drop.createdAt) >= drops.service.tuning.Drops.Lifetime+time.Minute {
			delete(drops.drops, id)
		}
	}
	for member, claimedAt := range drops.rainClaims {
		if now.Sub(claimedAt) >= drops.service.tuning.Drops.GiftRain.ClaimCooldown {
			delete(drops.rainClaims, member)
		}
	}
}

type GiftDropClaim struct {
	Member Member
	DropID string
	// Slot is the 1-based button of a Random Active Drop. Zero means the interface has no buttons.
	Slot int
}

type GiftDropAward struct {
	Kind   core.GiftDropKind
	Rarity core.GiftRarity
	Coal   bool
	// Taken has one bit per opened slot of a Random Active Drop.
	Taken uint8
	// Finished is set when this claim ended the drop, with the full list of collections.
	Finished *DropSummary
}

// Close ends a drop and returns who collected what. It reports false when the drop
// was already closed, so only one caller announces the end.
func (drops *ActivityDrops) Close(dropID string) (DropSummary, bool) {
	if drops == nil {
		return DropSummary{}, false
	}
	drops.mu.Lock()
	defer drops.mu.Unlock()
	return drops.closeLocked(dropID)
}

func (drops *ActivityDrops) closeLocked(dropID string) (DropSummary, bool) {
	drop, ok := drops.drops[dropID]
	if !ok {
		return DropSummary{}, false
	}
	delete(drops.drops, dropID)
	return DropSummary{Kind: drop.Kind, Claims: drop.results}, true
}

func (drops *ActivityDrops) Claim(ctx context.Context, claim GiftDropClaim) (GiftDropAward, error) {
	if drops == nil || claim.Member.ServerID == "" || claim.Member.UserID == "" || claim.DropID == "" {
		return GiftDropAward{}, ErrDropUnavailable
	}
	drops.mu.Lock()
	defer drops.mu.Unlock()
	drop, ok := drops.drops[claim.DropID]
	if !ok {
		return GiftDropAward{}, ErrDropUnavailable
	}
	now := drops.service.now().UTC()
	tuning := drops.service.tuning.Drops
	if now.Sub(drop.createdAt) >= tuning.Lifetime {
		return GiftDropAward{}, ErrDropUnavailable
	}
	slotBit := uint8(0)
	if claim.Slot >= 1 && claim.Slot <= core.GiftDropSlots {
		slotBit = 1 << (claim.Slot - 1)
	}
	if drop.Kind == core.RandomActiveDrop {
		if !drop.claims[claim.Member.UserID].IsZero() {
			return GiftDropAward{}, ErrDropCollected
		}
		if drop.taken&slotBit != 0 {
			return GiftDropAward{}, ErrDropSlotTaken
		}
	}
	if drop.Kind == core.GiftRainDrop && now.Sub(drops.rainClaims[claim.Member]) < tuning.GiftRain.ClaimCooldown {
		return GiftDropAward{}, ErrDropCooldown
	}
	coalRoll, rarityRoll := 0, drops.service.random(core.GiftRarityRolls)
	if drop.Kind == core.GiftRainDrop {
		coalRoll, rarityRoll = rarityRoll, drops.service.random(core.GiftRarityRolls)
	}
	for ctx.Err() == nil {
		record, err := drops.service.gifts.Load(ctx, claim.Member)
		if err != nil {
			return GiftDropAward{}, fmt.Errorf("load drop inventory: %w", err)
		}
		var next core.GiftRecord
		result := DropClaimRecord{UserID: claim.Member.UserID}
		switch {
		case drop.Kind == core.GiftRainDrop && tuning.RainPrizeIsCoal(coalRoll):
			result.Coal = true
			next, err = core.AwardCoal(record)
		case drop.Kind == core.GiftRainDrop:
			result.Rarity = drops.service.tuning.Gifts.DailyGiftRarity(rarityRoll)
			next, err = core.AwardGift(record, result.Rarity)
		default:
			result.Rarity = drops.service.tuning.Gifts.ActiveDropRarity(rarityRoll, record.Aura)
			next, err = core.AwardGift(record, result.Rarity)
		}
		if err != nil {
			return GiftDropAward{}, err
		}
		saved, err := drops.service.gifts.Save(ctx, GiftUpdate{Member: claim.Member, Previous: record, Next: next, Reason: ReasonDrop})
		if err != nil {
			return GiftDropAward{}, fmt.Errorf("save drop inventory: %w", err)
		}
		if saved {
			drop.claims[claim.Member.UserID] = now
			drop.taken |= slotBit
			drop.results = append(drop.results, result)
			drops.drops[drop.ID] = drop
			award := GiftDropAward{Kind: drop.Kind, Rarity: result.Rarity, Coal: result.Coal, Taken: drop.taken}
			if drop.Kind == core.GiftRainDrop {
				drops.rainClaims[claim.Member] = now
				if len(drop.results) >= tuning.GiftRain.ClaimsToEnd {
					summary, _ := drops.closeLocked(drop.ID)
					award.Finished = &summary
				}
			}
			return award, nil
		}
	}
	return GiftDropAward{}, ctx.Err()
}
