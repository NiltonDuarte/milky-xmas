package core

import (
	"cmp"
	"slices"
)

type GrinchDamage struct {
	Hundredths int
	Defenses   int
}

type GrinchScore struct {
	Member string
	Points float64
}

func (g Grinch) Leaderboard(rules GrinchRules) []GrinchScore {
	if g.DamageDealt == nil {
		return nil
	}

	players := g.defensePlayerCount()
	type entry struct {
		member    string
		numerator int
	}
	entries := make([]entry, 0, len(g.Joined))
	for member := range g.Joined {
		damage := g.DamageDealt[member]
		entries = append(entries, entry{member, damage.Hundredths*players + damage.Defenses*rules.DefenseTotal()})
	}
	slices.SortFunc(entries, func(a, b entry) int {
		if order := cmp.Compare(b.numerator, a.numerator); order != 0 {
			return order
		}
		return cmp.Compare(a.member, b.member)
	})
	scores := make([]GrinchScore, 0, len(entries))
	for _, ranked := range entries {
		scores = append(scores, GrinchScore{ranked.member, float64(ranked.numerator) / (100 * float64(players))})
	}
	return scores
}

func (g Grinch) GiftLeaderboard() []GrinchScore {
	if g.DamageDealt == nil {
		return nil
	}
	scores := make([]GrinchScore, 0, len(g.Joined))
	for member := range g.Joined {
		scores = append(scores, GrinchScore{Member: member, Points: float64(g.Gifts[member])})
	}
	slices.SortFunc(scores, func(a, b GrinchScore) int {
		if order := cmp.Compare(b.Points, a.Points); order != 0 {
			return order
		}
		return cmp.Compare(a.Member, b.Member)
	})
	return scores
}
