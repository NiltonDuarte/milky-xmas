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
	// Roll is a random number for the Grinch's lines. See GrinchEvent.
	Roll int
}

// Grinch uses hundredths of HP for attack damage. Defense HP derives from the
// exact number of successful defenses and the participant count.
type Grinch struct {
	ID        string
	MessageID string
	Phase     GrinchPhase
	StartedAt time.Time
	// CountdownAt is when the first player pressed Start. Registration closes RegistrationDuration
	// later. It is zero while nobody has pressed Start.
	CountdownAt    time.Time
	TimedAt        time.Time
	EndedAt        time.Time
	HP             int
	Board          uint8
	FinalBoard     uint8
	FinalRound     int
	Gifts          map[string]int
	Snowballs      map[string]int // Snowballs each player collected with the final ❄️ button.
	PlayerFeed     []GrinchEvent  // The latest attacks, for the players' text area.
	BossFeed       []GrinchEvent  // The latest Grinch lines and what he did, for the boss' text area.
	Defenses       int
	DefenseRound   int
	DefensePlayers int
	Joined         map[string]bool
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
	return max(1, g.DefensePlayers)
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
	case GrinchSnowballs:
		// The group has this long to bring the Grinch down to the snowball milestone.
		return g.StartedAt.Add(rules.SnowballsDuration)
	case GrinchDefense, GrinchAttack, GrinchFinalAttack:
		return g.TimedAt.Add(rules.FightDuration)
	default:
		return time.Time{}
	}
}

// LobbyDeadline is when registration closes and combat starts or the fight is cancelled. The
// countdown begins when the first player presses Start. Until then the fight waits for StartWindow.
func (g Grinch) LobbyDeadline(rules GrinchRules) time.Time {
	switch {
	case !g.CountdownAt.IsZero():
		return g.CountdownAt.Add(rules.RegistrationDuration)
	default:
		return g.StartedAt.Add(rules.StartWindow)
	}
}

func (g *Grinch) Advance(rules GrinchRules, now time.Time) bool {
	if g.ID == "" || !g.EndedAt.IsZero() {
		return false
	}
	started := g.Phase == GrinchLobby
	if started {
		deadline := g.LobbyDeadline(rules)
		if now.Before(deadline) {
			return false
		}
		g.Dirty = true
		if len(g.Joined) == 0 {
			g.Phase, g.EndedAt = GrinchCancelled, deadline
			return true
		}
		g.Phase, g.StartedAt = GrinchSnowballs, deadline
		g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchSnowballs, Roll: int(deadline.UnixMilli())})
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
	for member := range g.Joined {
		if g.PlayerDamage[member] < 100 {
			return true
		}
	}
	return false
}

// Move applies a player's action. Anyone can take part, even after the fight has started: the first
// action of a player that is accepted puts them in the fight, and a refused action does not.
func (g *Grinch) Move(rules GrinchRules, move GrinchMove) string {
	if move.Phase == GrinchLobby || g.Phase == GrinchLobby || g.Joined == nil || g.Joined[move.Member] {
		return g.move(rules, move)
	}
	g.Joined[move.Member] = true
	reply := g.move(rules, move)
	if !g.Accepted[move.RequestID] {
		delete(g.Joined, move.Member)
	}
	return reply
}

func (g *Grinch) move(rules GrinchRules, move GrinchMove) string {
	if g.Accepted[move.RequestID] {
		return "This action was already counted."
	}
	deadline := g.Deadline(rules)
	if !g.EndedAt.IsZero() || (!deadline.IsZero() && !move.At.Before(deadline)) {
		return "This part of the fight has ended."
	}
	if move.Phase == GrinchLobby {
		if g.Joined == nil || move.Slot != 0 || (g.Phase == GrinchLobby && !move.At.Before(g.LobbyDeadline(rules))) {
			return "This fight is no longer accepting players."
		}
		if g.Joined[move.Member] {
			return "✅ You already pressed Start."
		}
		g.Joined[move.Member], g.Accepted[move.RequestID], g.Dirty = true, true, true
		if g.Phase == GrinchLobby && g.CountdownAt.IsZero() {
			g.CountdownAt = move.At // The first Start press begins the countdown.
		}
		if g.Phase == GrinchLobby && len(g.Joined) >= rules.EarlyStartPlayers {
			g.Phase, g.StartedAt = GrinchSnowballs, move.At
			g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchSnowballs, Roll: move.Roll})
		}
		if g.Phase == GrinchLobby {
			return "✅ You pressed Start! The fight begins soon."
		}
		return "✅ You're in the fight! You can attack now."
	}
	if g.Phase != move.Phase || move.At.Before(g.StartedAt) {
		return "This part of the fight has ended."
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
		applied := min(rules.SnowballDamage, g.HP-rules.SnowballsBossHP)
		g.HP -= applied
		damage.Hundredths = applied
		reply = fmt.Sprintf("Snowball hit for %s HP.", FormatHundredths(applied))
		if g.HP == rules.SnowballsBossHP {
			g.DefensePlayers = len(g.Joined)
			g.Phase, g.Board, g.TimedAt, g.DefenseRound = GrinchDefense, move.NextBoard, move.At, 1
			g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchDefense, Roll: move.Roll})
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
			g.bossEvent(GrinchEvent{Kind: EventMiss, Member: move.Member, Amount: rules.DefenseMissHPLoss, Roll: move.Roll})
			if g.PlayerDamage[move.Member] >= 100 && !g.hasSurvivor() {
				g.end(rules, GrinchDefeat, move.At)
				g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchDefeat, Roll: move.Roll})
			}
		} else {
			g.Defenses++
			damage.Defenses = 1
			reply = "Safe defense! The Grinch lost HP."
			if g.Defenses == g.defensePlayerCount() {
				g.Phase, g.HP, g.Board = GrinchAttack, rules.DefenseBossHP, move.NextBoard
				g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchAttack, Roll: move.Roll})
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
			g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchFinalAttack, Roll: move.Roll})
		}
	case GrinchFinalAttack:
		if move.FinalRound != g.FinalRound {
			return "This row changed. Choose a button on the current row."
		}
		if move.Slot < 0 || move.Slot > GrinchSnowballSlot || (g.FinalRound == 0 && move.Slot != 0 && move.Slot != GrinchSnowballSlot) {
			return "Choose a button on the final row."
		}
		if move.Slot == GrinchSnowballSlot {
			if g.Snowballs == nil {
				g.Snowballs = make(map[string]int)
			}
			g.Snowballs[move.Member] += int(rules.FinalSnowballReward)
			g.bossEvent(GrinchEvent{Kind: EventSnowball, Member: move.Member, Amount: int(rules.FinalSnowballReward), Roll: move.Roll})
			reply = rules.SnowballEarned()
		} else if g.FinalBoard&(1<<move.Slot) != 0 {
			if move.GiftPoints < rules.GiftPointsMin || move.GiftPoints > rules.GiftPointsMax {
				return "Could not award gift points."
			}
			if g.Gifts == nil {
				g.Gifts = make(map[string]int)
			}
			g.Gifts[move.Member] += move.GiftPoints
			g.bossEvent(GrinchEvent{Kind: EventGift, Member: move.Member, Amount: move.GiftPoints, Roll: move.Roll})
			reply = fmt.Sprintf("🎁 You earned %d gift points!", move.GiftPoints)
		} else {
			applied := min(g.HP, rules.FinalHitDamage)
			g.HP -= applied
			damage.Hundredths, damage.Points = applied, rules.FinalHitScore
			reply = fmt.Sprintf("Final attack hit for %s HP.", FormatHundredths(applied))
			if g.HP == 0 {
				g.end(rules, GrinchWon, move.At)
			}
		}
		g.FinalBoard, g.FinalRound = move.FinalBoard, g.FinalRound+1
	default:
		return "This part of the fight has ended."
	}
	switch {
	case damage.Hundredths != 0:
		g.playerEvent(GrinchEvent{Kind: EventHit, Member: move.Member, Damage: float64(damage.Hundredths) / 100})
	case damage.Defenses != 0:
		g.playerEvent(GrinchEvent{Kind: EventDefend, Member: move.Member, Roll: move.Roll,
			Damage: float64(damage.Defenses*rules.DefenseTotal()) / (100 * float64(g.defensePlayerCount()))})
	}
	if g.DamageDealt != nil && (damage.Hundredths != 0 || damage.Defenses != 0) {
		total := g.DamageDealt[move.Member]
		total.Hundredths += damage.Hundredths
		total.Defenses += damage.Defenses
		total.Points += damage.Points // Only the final attack awards points.
		g.DamageDealt[move.Member] = total
	}
	g.Accepted[move.RequestID] = true
	g.Dirty = true
	return reply
}
