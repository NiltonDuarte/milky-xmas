package app

import (
	"context"
	"fmt"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

type Member struct {
	ServerID string
	UserID   string
}

type GiftStore interface {
	Load(context.Context, Member) (core.GiftRecord, error)
	// Save replaces the record only if the stored total equals the observed total.
	// A missing record has a total of zero.
	Save(context.Context, Member, int64, core.GiftRecord) (bool, error)
}

type GiftClaim struct {
	Collected bool
	Total     int64
	NextClaim time.Time
}

func NewService(store Store, now func() time.Time) Service {
	return Service{gifts: store, advent: store, now: now}
}

func (service Service) ClaimGift(ctx context.Context, member Member) (GiftClaim, error) {
	now := service.now().UTC()
	for ctx.Err() == nil {
		record, err := service.gifts.Load(ctx, member)
		if err != nil {
			return GiftClaim{}, fmt.Errorf("load gift total: %w", err)
		}
		updated, next, collected := record.Collect(now)
		if !collected {
			return GiftClaim{Total: record.Total, NextClaim: next}, nil
		}
		saved, err := service.gifts.Save(ctx, member, record.Total, updated)
		if err != nil {
			return GiftClaim{}, fmt.Errorf("save gift claim: %w", err)
		}
		if saved {
			return GiftClaim{Collected: true, Total: updated.Total, NextClaim: next}, nil
		}
	}
	return GiftClaim{}, ctx.Err()
}
