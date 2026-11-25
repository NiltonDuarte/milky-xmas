package core

import (
	"cmp"
	"slices"
)

// GrinchDamage is what a player did to the Grinch. Hundredths is the HP their hits removed and
// Defenses counts the safe defenses, which share the defense phase's HP between the players.
// Points are score points, in hundredths, and only the final attack awards them: each final hit
// gives FinalHitScore points, which can differ from the HP it removes. The snowball, defense, and
// attack phases award none. The points are kept for later use and no leaderboard shows them yet.
type GrinchDamage struct {
	Hundredths int
	Defenses   int
	Points     int
}

type GrinchScore struct {
	Member string
	Points float64
}

// WithPoints keeps the players who scored above zero, in the same order. The final leaderboards
// list only them.
func WithPoints(scores []GrinchScore) []GrinchScore {
	kept := make([]GrinchScore, 0, len(scores))
	for _, score := range scores {
		if score.Points > 0 {
			kept = append(kept, score)
		}
	}
	return kept
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

func (g Grinch) GiftLeaderboard() []GrinchScore { return g.countLeaderboard(g.Gifts) }

// SnowballLeaderboard ranks the players who joined by the Snowballs they collected.
func (g Grinch) SnowballLeaderboard() []GrinchScore { return g.countLeaderboard(g.Snowballs) }

func (g Grinch) countLeaderboard(counts map[string]int) []GrinchScore {
	if g.DamageDealt == nil {
		return nil
	}
	scores := make([]GrinchScore, 0, len(g.Joined))
	for member := range g.Joined {
		scores = append(scores, GrinchScore{Member: member, Points: float64(counts[member])})
	}
	slices.SortFunc(scores, func(a, b GrinchScore) int {
		if order := cmp.Compare(b.Points, a.Points); order != 0 {
			return order
		}
		return cmp.Compare(a.Member, b.Member)
	})
	return scores
}
