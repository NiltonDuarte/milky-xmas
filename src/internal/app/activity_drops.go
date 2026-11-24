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

const (
	GiftRainClaimCooldown = time.Minute
	GiftDropLifetime      = 10 * time.Minute
)

var (
	ErrDropUnavailable = errors.New("gift drop is no longer available")
	ErrDropCollected   = errors.New("gift drop already collected")
	ErrDropCooldown    = errors.New("gift rain collection is cooling down")
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
	if activity.target == 0 {
		activity.target = core.NextGiftRainThreshold(drops.service.random(core.GiftRainMessageRange))
	}
	activity.messages++
	if activity.messages >= activity.target {
		activity.messages = 0
		activity.target = core.NextGiftRainThreshold(drops.service.random(core.GiftRainMessageRange))
		drops.activity[key] = activity
		return drops.createLocked(core.GiftRainDrop)
	}
	drops.activity[key] = activity
	if core.RandomActiveDropEligible(drops.service.random(core.RandomActiveDropRolls)) {
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
		if now.Sub(drop.createdAt) >= GiftDropLifetime {
			delete(drops.drops, id)
		}
	}
	for member, claimedAt := range drops.rainClaims {
		if now.Sub(claimedAt) >= GiftRainClaimCooldown {
			delete(drops.rainClaims, member)
		}
	}
}

type GiftDropClaim struct {
	Member Member
	DropID string
}

type GiftDropAward struct {
	Kind   core.GiftDropKind
	Rarity core.GiftRarity
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
	if now.Sub(drop.createdAt) >= GiftDropLifetime {
		delete(drops.drops, drop.ID)
		return GiftDropAward{}, ErrDropUnavailable
	}
	if drop.Kind == core.RandomActiveDrop && !drop.claims[claim.Member.UserID].IsZero() {
		return GiftDropAward{}, ErrDropCollected
	}
	if drop.Kind == core.GiftRainDrop && now.Sub(drops.rainClaims[claim.Member]) < GiftRainClaimCooldown {
		return GiftDropAward{}, ErrDropCooldown
	}
	rarity := core.DailyGiftRarity(drops.service.random(core.GiftRarityRolls))
	for ctx.Err() == nil {
		record, err := drops.service.gifts.Load(ctx, claim.Member)
		if err != nil {
			return GiftDropAward{}, fmt.Errorf("load drop inventory: %w", err)
		}
		next, err := core.AwardGift(record, rarity)
		if err != nil {
			return GiftDropAward{}, err
		}
		saved, err := drops.service.gifts.Save(ctx, GiftUpdate{Member: claim.Member, Previous: record, Next: next})
		if err != nil {
			return GiftDropAward{}, fmt.Errorf("save drop inventory: %w", err)
		}
		if saved {
			drop.claims[claim.Member.UserID] = now
			drops.drops[drop.ID] = drop
			if drop.Kind == core.GiftRainDrop {
				drops.rainClaims[claim.Member] = now
			}
			return GiftDropAward{Kind: drop.Kind, Rarity: rarity}, nil
		}
	}
	return GiftDropAward{}, ctx.Err()
}
