package core

import (
	"cmp"
	"slices"
)

// PlayerActivity is what one player did during the event.
type PlayerActivity struct {
	NormalGiftsSent, RareGiftsSent, EpicGiftsSent int64
	CoalSent, SnowballsThrown, YellowThrown       int64
	DropsClaimed, Waterings                       int64
}

// GiftsSent counts the gifts of every rarity the player sent.
func (activity PlayerActivity) GiftsSent() int64 {
	return activity.NormalGiftsSent + activity.RareGiftsSent + activity.EpicGiftsSent
}

// GoodPoints adds up every good deed. Unlike the aura, bad deeds do not lower it and it has no
// limit. waterPoints is what one watering is worth.
func (activity PlayerActivity) GoodPoints(waterPoints int) int64 {
	return activity.NormalGiftsSent*NormalGiftPoints + activity.RareGiftsSent*RareGiftPoints +
		activity.EpicGiftsSent*EpicGiftPoints + activity.Waterings*int64(waterPoints)
}

// BadPoints adds up every bad deed, in the same way as GoodPoints.
func (activity PlayerActivity) BadPoints() int64 {
	return activity.CoalSent*CoalPoints + activity.YellowThrown*YellowSnowballPoints
}

// FightTotals counts the Grinch fights that were fought to an end.
type FightTotals struct {
	Won, Lost int
}

// EventStats is everything the event's statistics are drawn from.
type EventStats struct {
	Players     map[string]PlayerActivity
	Fights      FightTotals
	TreeLevel   int
	WaterPoints int // The good deed points of one watering.
}

// Total adds up value over every player.
func (stats EventStats) Total(value func(PlayerActivity) int64) int64 {
	var total int64
	for _, activity := range stats.Players {
		total += value(activity)
	}
	return total
}

// StatEntry is a player's place on a leaderboard.
type StatEntry struct {
	Member string
	Count  int64
}

// Ranking lists the players whose value is above zero, highest first, ties broken by ID.
func (stats EventStats) Ranking(value func(PlayerActivity) int64) []StatEntry {
	var entries []StatEntry
	for member, activity := range stats.Players {
		if count := value(activity); count > 0 {
			entries = append(entries, StatEntry{member, count})
		}
	}
	slices.SortFunc(entries, func(a, b StatEntry) int {
		if order := cmp.Compare(b.Count, a.Count); order != 0 {
			return order
		}
		return cmp.Compare(a.Member, b.Member)
	})
	return entries
}
