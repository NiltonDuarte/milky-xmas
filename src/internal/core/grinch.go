package core

import (
	"fmt"
	"time"
)

// GrinchSnowballSlot is the fixed fifth button of the final attack row.
const GrinchSnowballSlot = 4

// SnowballEarned is the reply for the final row's Snowball button.
func (rules GrinchRules) SnowballEarned() string {
	if rules.FinalSnowballReward == 1 {
		return "❄️ You earned 1 Snowball!"
	}
	return fmt.Sprintf("❄️ You earned %d Snowballs!", rules.FinalSnowballReward)
}

// SnowflakeHit is the reply for a missed defense.
func (rules GrinchRules) SnowflakeHit() string {
	return fmt.Sprintf("Snowflake! You lost %d%% of your maximum HP.", rules.DefenseMissHPLoss)
}

type GrinchPhase string

const (
	GrinchLobby       GrinchPhase = "lobby"
	GrinchCancelled   GrinchPhase = "cancelled"
	GrinchPending     GrinchPhase = "pending"
	GrinchSnowballs   GrinchPhase = "snowballs"
	GrinchDefense     GrinchPhase = "defense"
	GrinchAttack      GrinchPhase = "attack"
	GrinchFinalAttack GrinchPhase = "final"
	GrinchWaiting     GrinchPhase = "waiting"
	GrinchWon         GrinchPhase = "won"
	GrinchDefeat      GrinchPhase = "defeat"
	GrinchEscaped     GrinchPhase = "escaped"
)

type GrinchMove struct {
	Phase        GrinchPhase
	Slot         int
	Member       string
	RequestID    string
	At           time.Time
	NextBoard    uint8
	DefenseRound int
	FinalBoard   uint8
	FinalRound   int
	GiftPoints   int
}

// Grinch uses hundredths of HP for attack damage. Defense HP derives from the
// exact number of successful defenses and the participant count.
type Grinch struct {
	ID             string
	MessageID      string
	Phase          GrinchPhase
	StartedAt      time.Time
	TimedAt        time.Time
	EndedAt        time.Time
	HP             int
	Board          uint8
	FinalBoard     uint8
	FinalRound     int
	Gifts          map[string]int
	Defenses       int
	DefenseRound   int
	DefensePlayers int
	Joined         map[string]bool
	Participants   map[string]bool
	PlayerDamage   map[string]int
	DamageDealt    map[string]GrinchDamage
	Accepted       map[string]bool
	Dirty          bool
}

func (g Grinch) Health(rules GrinchRules) float64 {
	if g.Phase == GrinchDefense {
		return float64(rules.SnowballsBossHP)/100 - float64(g.Defenses*rules.DefenseTotal())/(100*float64(g.defensePlayerCount()))
	}
	return float64(g.HP) / 100
}

func (g Grinch) defensePlayerCount() int {
	if g.DefensePlayers > 0 {
		return g.DefensePlayers
	}
	// Preserve the scale of a defense phase already in progress when upgraded.
	return max(1, len(g.Participants))
}

func (g Grinch) CanSpawn(rules GrinchRules, now, activity time.Time, roll int, manual bool) bool {
	if g.ID != "" && g.EndedAt.IsZero() {
		return false
	}
	if manual {
		return true
	}
	return !activity.IsZero() && !activity.After(now) && now.Sub(activity) <= rules.ActivityWindow &&
		(g.EndedAt.IsZero() || now.Sub(g.EndedAt) >= rules.CooldownAfterFight) && roll >= 0 && roll < rules.SpawnChancePercent
}

func (g Grinch) Deadline(rules GrinchRules) time.Time {
	switch g.Phase {
	case GrinchPending:
		return g.StartedAt.Add(rules.FightDuration)
	case GrinchDefense, GrinchAttack, GrinchFinalAttack:
		if g.TimedAt.IsZero() {
			// Older encounters started their countdown when combat began.
			return g.StartedAt.Add(rules.FightDuration)
		}
		return g.TimedAt.Add(rules.FightDuration)
	default:
		return time.Time{}
	}
}

func (g *Grinch) Advance(rules GrinchRules, now time.Time) bool {
	if g.ID == "" || !g.EndedAt.IsZero() {
		return false
	}
	started := g.Phase == GrinchLobby
	if started {
		deadline := g.StartedAt.Add(rules.RegistrationDuration)
		if now.Before(deadline) {
			return false
		}
		g.Dirty = true
		if len(g.Joined) == 0 {
			g.Phase, g.EndedAt = GrinchCancelled, deadline
			return true
		}
		g.Phase, g.StartedAt = GrinchSnowballs, deadline
	}
	if g.Phase == GrinchWaiting {
		// Resume the snowball phase for older fights paused by its old time limit.
		g.Phase, g.Dirty = GrinchSnowballs, true
		return true
	}
	if deadline := g.Deadline(rules); !deadline.IsZero() && !now.Before(deadline) {
		g.end(rules, GrinchEscaped, deadline)
		return true
	}
	return started
}

func (g *Grinch) end(rules GrinchRules, phase GrinchPhase, at time.Time) {
	if g.Phase == GrinchDefense {
		g.HP = rules.SnowballsBossHP - g.Defenses*rules.DefenseTotal()/g.defensePlayerCount()
	}
	g.Phase, g.EndedAt, g.Dirty = phase, at, true
}

func (g Grinch) hasSurvivor() bool {
	roster := g.Joined
	if roster == nil {
		roster = g.Participants
	}
	for member := range roster {
		if g.PlayerDamage[member] < 100 {
			return true
		}
	}
	return false
}

func (g Grinch) CanJoin(rules GrinchRules, member string, at time.Time) bool {
	deadline := g.Deadline(rules)
	return g.Joined != nil && !g.Joined[member] && g.EndedAt.IsZero() &&
		(g.Phase != GrinchLobby || at.Before(g.StartedAt.Add(rules.RegistrationDuration))) &&
		(deadline.IsZero() || at.Before(deadline))
}

func (g *Grinch) Move(rules GrinchRules, move GrinchMove) string {
	if g.Accepted[move.RequestID] {
		return "This action was already counted."
	}
	deadline := g.Deadline(rules)
	if !g.EndedAt.IsZero() || (!deadline.IsZero() && !move.At.Before(deadline)) {
		return "This part of the fight has ended."
	}
	if move.Phase == GrinchLobby {
		if g.Joined == nil || move.Slot != 0 || (g.Phase == GrinchLobby && !move.At.Before(g.StartedAt.Add(rules.RegistrationDuration))) {
			return "This fight is no longer accepting players."
		}
		if g.Joined[move.Member] {
			return "✅ You already joined this fight."
		}
		g.Joined[move.Member], g.Accepted[move.RequestID], g.Dirty = true, true, true
		if g.Phase == GrinchLobby && len(g.Joined) >= rules.EarlyStartPlayers {
			g.Phase, g.StartedAt = GrinchSnowballs, move.At
		}
		if g.Phase == GrinchLobby {
			return "✅ You joined the Grinch fight! Wait for combat to start."
		}
		return "✅ You joined the Grinch fight! You can attack now."
	}
	if g.Phase != move.Phase || move.At.Before(g.StartedAt) {
		return "This part of the fight has ended."
	}
	// Older persisted fights began before registration was introduced.
	if g.Joined != nil && !g.Joined[move.Member] {
		return "Join the fight using 🎅 to take part."
	}
	if g.PlayerDamage[move.Member] >= 100 {
		return "You have no HP left in this fight."
	}
	var reply string
	damage := GrinchDamage{}
	switch g.Phase {
	case GrinchSnowballs:
		if move.Slot != 0 {
			return "The snowball phase has ended."
		}
		if g.Joined == nil {
			g.Participants[move.Member] = true
		}
		applied := min(rules.SnowballDamage, g.HP-rules.SnowballsBossHP)
		g.HP -= applied
		damage.Hundredths = applied
		reply = fmt.Sprintf("Snowball hit for %s HP.", FormatHundredths(applied))
		if g.HP == rules.SnowballsBossHP {
			g.DefensePlayers = len(g.Joined)
			if g.DefensePlayers == 0 {
				g.DefensePlayers = g.defensePlayerCount()
			}
			g.Phase, g.Board, g.TimedAt, g.DefenseRound = GrinchDefense, move.NextBoard, move.At, 1
		}
	case GrinchDefense:
		if move.DefenseRound != g.DefenseRound {
			return "The defense board changed. Choose a button on the current board."
		}
		if move.Slot < 0 || move.Slot >= 8 {
			return "That defense tile is unavailable."
		}
		if g.Board&(1<<move.Slot) == 0 {
			g.PlayerDamage[move.Member] += rules.DefenseMissHPLoss
			reply = rules.SnowflakeHit()
			if g.PlayerDamage[move.Member] >= 100 && !g.hasSurvivor() {
				g.end(rules, GrinchDefeat, move.At)
			}
		} else {
			g.Defenses++
			damage.Defenses = 1
			reply = "Safe defense! The Grinch lost HP."
			if g.Defenses == g.defensePlayerCount() {
				g.Phase, g.HP, g.Board = GrinchAttack, rules.DefenseBossHP, move.NextBoard
			}
		}
		if g.Phase == GrinchDefense {
			next := move.NextBoard
			if next == 0 || next == 255 || next == g.Board {
				next = g.Board%254 + 1
			}
			g.Board = next
			g.DefenseRound++
		}
	case GrinchAttack:
		if move.Slot < 0 || move.Slot >= 8 {
			return "That attack tile is unavailable."
		}
		hit := rules.AttackDamage
		if g.Board&(1<<move.Slot) != 0 {
			hit = rules.AttackHeavyDamage
		}
		damageApplied := min(g.HP-rules.AttackBossHP, hit)
		g.HP -= damageApplied
		damage.Hundredths = damageApplied
		reply = "Attack hit!"
		if g.HP == rules.AttackBossHP {
			g.Phase, g.FinalBoard, g.FinalRound = GrinchFinalAttack, move.FinalBoard, 1
		}
	case GrinchFinalAttack:
		if move.FinalRound != g.FinalRound {
			return "This row changed. Choose a button on the current row."
		}
		if move.Slot < 0 || move.Slot > GrinchSnowballSlot || (g.FinalRound == 0 && move.Slot != 0 && move.Slot != GrinchSnowballSlot) {
			return "Choose a button on the final row."
		}
		if move.Slot == GrinchSnowballSlot {
			reply = rules.SnowballEarned()
		} else if g.FinalBoard&(1<<move.Slot) != 0 {
			if move.GiftPoints < rules.GiftPointsMin || move.GiftPoints > rules.GiftPointsMax {
				return "Could not award gift points."
			}
			if g.Gifts == nil {
				g.Gifts = make(map[string]int)
			}
			g.Gifts[move.Member] += move.GiftPoints
			reply = fmt.Sprintf("🎁 You earned %d gift points!", move.GiftPoints)
		} else {
			g.HP = max(0, g.HP-rules.FinalHitDamage)
			damage.Hundredths = rules.FinalHitScore
			reply = fmt.Sprintf("Final attack hit! You earned %s damage points.", FormatHundredths(rules.FinalHitScore))
			if g.HP == 0 {
				g.end(rules, GrinchWon, move.At)
			}
		}
		g.FinalBoard, g.FinalRound = move.FinalBoard, g.FinalRound+1
	default:
		return "This part of the fight has ended."
	}
	if g.DamageDealt != nil && (damage.Hundredths != 0 || damage.Defenses != 0) {
		total := g.DamageDealt[move.Member]
		total.Hundredths += damage.Hundredths
		total.Defenses += damage.Defenses
		g.DamageDealt[move.Member] = total
	}
	g.Accepted[move.RequestID] = true
	g.Dirty = true
	return reply
}
