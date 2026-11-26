package app

import (
	"context"
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// HistoryEntry is one saved change to a player's inventory.
type HistoryEntry struct {
	UserID string
	Reason Reason
	Delta  core.Inventory // What the change added; negative values were taken away.
}

// StatsStore reads the records the statistics are built from.
type StatsStore interface {
	PlayerHistory(context.Context, string) ([]HistoryEntry, error)
	// FightResults returns how every finished Grinch fight of a server ended.
	FightResults(context.Context, string) ([]core.GrinchPhase, error)
}

// Stats gathers the server's event statistics. Waterings and the tree level come from tree.
func (service Service) Stats(ctx context.Context, server string, tree *Tree) (core.EventStats, error) {
	if server == "" || tree == nil {
		return core.EventStats{}, ErrInvalidRecipient
	}
	history, err := service.stats.PlayerHistory(ctx, server)
	if err != nil {
		return core.EventStats{}, fmt.Errorf("load player history: %w", err)
	}
	results, err := service.stats.FightResults(ctx, server)
	if err != nil {
		return core.EventStats{}, fmt.Errorf("load fight results: %w", err)
	}
	current, err := tree.View(ctx, server)
	if err != nil {
		return core.EventStats{}, fmt.Errorf("load tree: %w", err)
	}
	stats := core.EventStats{
		Players: make(map[string]core.PlayerActivity), TreeLevel: current.Level, WaterPoints: tree.rules.WaterAura,
	}
	for _, entry := range history {
		activity := stats.Players[entry.UserID]
		// Only what a player gave away counts as sent or thrown; the receiving side is positive.
		given := func(delta int64) int64 { return max(0, -delta) }
		switch entry.Reason {
		case ReasonSendGift:
			activity.NormalGiftsSent += given(entry.Delta.Normal)
			activity.RareGiftsSent += given(entry.Delta.Rare)
			activity.EpicGiftsSent += given(entry.Delta.Epic)
		case ReasonSendCoal:
			activity.CoalSent += given(entry.Delta.Coal)
		case ReasonThrowSnowball:
			activity.SnowballsThrown += given(entry.Delta.Snowballs)
			activity.YellowThrown += given(entry.Delta.YellowSnowballs)
		case ReasonDrop:
			activity.DropsClaimed++
		}
		stats.Players[entry.UserID] = activity
	}
	for user, waterings := range current.Waters {
		activity := stats.Players[user]
		activity.Waterings = int64(waterings)
		stats.Players[user] = activity
	}
	for _, phase := range results {
		if phase == core.GrinchWon {
			stats.Fights.Won++
		} else {
			stats.Fights.Lost++
		}
	}
	return stats, nil
}
