package core

import (
	"fmt"
	"time"
)

const (
	GrinchJoinDuration = time.Minute
	GrinchDuration     = time.Minute
	GrinchSnowflakeHit = "Snowflake! You lost 10% of your maximum HP."
)

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

func (g Grinch) Health() float64 {
	if g.Phase == GrinchDefense {
		return 75 - float64(g.Defenses*30)/float64(g.defensePlayerCount())
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

func (g Grinch) CanSpawn(now, activity time.Time, roll int, manual bool) bool {
	if g.ID != "" && g.EndedAt.IsZero() {
		return false
	}
	if manual {
		return true
	}
	return !activity.IsZero() && !activity.After(now) && now.Sub(activity) <= 10*time.Minute &&
		(g.EndedAt.IsZero() || now.Sub(g.EndedAt) >= 30*time.Minute) && roll >= 0 && roll < 10
}

func (g Grinch) Deadline() time.Time {
	switch g.Phase {
	case GrinchPending:
		return g.StartedAt.Add(GrinchDuration)
	case GrinchDefense, GrinchAttack, GrinchFinalAttack:
		if g.TimedAt.IsZero() {
			// Older encounters started their countdown when combat began.
			return g.StartedAt.Add(GrinchDuration)
		}
		return g.TimedAt.Add(GrinchDuration)
	default:
		return time.Time{}
	}
}

func (g *Grinch) Advance(now time.Time) bool {
	if g.ID == "" || !g.EndedAt.IsZero() {
		return false
	}
	started := g.Phase == GrinchLobby
	if started {
		deadline := g.StartedAt.Add(GrinchJoinDuration)
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
	if deadline := g.Deadline(); !deadline.IsZero() && !now.Before(deadline) {
		g.end(GrinchEscaped, deadline)
		return true
	}
	return started
}

func (g *Grinch) end(phase GrinchPhase, at time.Time) {
	if g.Phase == GrinchDefense {
		g.HP = 7500 - g.Defenses*3000/g.defensePlayerCount()
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

func (g *Grinch) Move(move GrinchMove) string {
	if g.Accepted[move.RequestID] {
		return "This action was already counted."
	}
	deadline := g.Deadline()
	if !g.EndedAt.IsZero() || (!deadline.IsZero() && !move.At.Before(deadline)) {
		return "This part of the fight has ended."
	}
	if move.Phase == GrinchLobby {
		if g.Joined == nil || move.Slot != 0 || (g.Phase == GrinchLobby && !move.At.Before(g.StartedAt.Add(GrinchJoinDuration))) {
			return "This fight is no longer accepting players."
		}
		if g.Joined[move.Member] {
			return "✅ You already joined this fight."
		}
		g.Joined[move.Member], g.Accepted[move.RequestID], g.Dirty = true, true, true
		if g.Phase == GrinchLobby && len(g.Joined) == 3 {
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
		g.HP -= 100
		damage.Hundredths = 100
		reply = "Snowball hit for 1 HP."
		if g.HP == 7500 {
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
			g.PlayerDamage[move.Member] += 10
			reply = GrinchSnowflakeHit
			if g.PlayerDamage[move.Member] == 100 && !g.hasSurvivor() {
				g.end(GrinchDefeat, move.At)
			}
		} else {
			g.Defenses++
			damage.Defenses = 1
			reply = "Safe defense! The Grinch lost HP."
			if g.Defenses == g.defensePlayerCount() {
				g.Phase, g.HP, g.Board = GrinchAttack, 4500, move.NextBoard
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
		hit := 500
		if g.Board&(1<<move.Slot) != 0 {
			hit = 525
		}
		damageApplied := min(g.HP-500, hit)
		g.HP -= damageApplied
		damage.Hundredths = damageApplied
		reply = "Attack hit!"
		if g.HP == 500 {
			g.Phase, g.FinalBoard, g.FinalRound = GrinchFinalAttack, move.FinalBoard, 1
		}
	case GrinchFinalAttack:
		if move.FinalRound != g.FinalRound {
			return "This row changed. Choose a button on the current row."
		}
		if move.Slot < 0 || move.Slot >= 4 || (g.FinalRound == 0 && move.Slot != 0) {
			return "Choose a button on the final row."
		}
		if g.FinalBoard&(1<<move.Slot) != 0 {
			if move.GiftPoints < 1 || move.GiftPoints > 5 {
				return "Could not award gift points."
			}
			if g.Gifts == nil {
				g.Gifts = make(map[string]int)
			}
			g.Gifts[move.Member] += move.GiftPoints
			reply = fmt.Sprintf("🎁 You earned %d gift points!", move.GiftPoints)
		} else {
			g.HP = max(0, g.HP-100)
			damage.Hundredths = 200
			reply = "Final attack hit! You earned 2 damage points."
			if g.HP == 0 {
				g.end(GrinchWon, move.At)
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
