package core

import "time"

const (
	GrinchJoinDuration = 10 * time.Second
	GrinchDuration     = time.Minute
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
	GrinchEscaped     GrinchPhase = "escaped"
)

type GrinchMove struct {
	Phase     GrinchPhase
	Slot      int
	Member    string
	RequestID string
	At        time.Time
	NextBoard uint8
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
	Defenses       int
	DefensePlayers int
	Joined         map[string]bool
	Participants   map[string]bool
	PlayerDamage   map[string]int
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
		if g.Phase == GrinchDefense {
			g.HP = 7500 - g.Defenses*3000/g.defensePlayerCount()
		}
		g.Phase, g.EndedAt, g.Dirty = GrinchEscaped, deadline, true
		return true
	}
	return started
}

func (g *Grinch) Move(move GrinchMove) string {
	if g.Accepted[move.RequestID] {
		return "This action was already counted."
	}
	deadline := g.Deadline()
	if g.Phase != move.Phase || !g.EndedAt.IsZero() || move.At.Before(g.StartedAt) ||
		(!deadline.IsZero() && !move.At.Before(deadline)) {
		return "This part of the fight has ended."
	}
	if g.Phase == GrinchLobby {
		if move.Slot != 0 || !move.At.Before(g.StartedAt.Add(GrinchJoinDuration)) {
			return "Registration has closed."
		}
		if g.Joined[move.Member] {
			return "You already joined this fight."
		}
		g.Joined[move.Member], g.Accepted[move.RequestID], g.Dirty = true, true, true
		return "You joined the Grinch fight!"
	}
	// Older persisted fights began before registration was introduced.
	if g.Joined != nil && !g.Joined[move.Member] {
		return "Join the fight using 🎅 before combat starts."
	}
	if g.PlayerDamage[move.Member] >= 100 {
		return "You have no HP left in this fight."
	}
	var reply string
	switch g.Phase {
	case GrinchSnowballs:
		if move.Slot != 0 {
			return "The snowball phase has ended."
		}
		if g.Joined == nil {
			g.Participants[move.Member] = true
		}
		g.HP -= 100
		reply = "Snowball hit for 1 HP."
		if g.HP == 7500 {
			g.DefensePlayers = len(g.Joined)
			if g.DefensePlayers == 0 {
				g.DefensePlayers = g.defensePlayerCount()
			}
			g.Phase, g.Board, g.TimedAt = GrinchDefense, move.NextBoard, move.At
		}
	case GrinchDefense:
		if move.Slot < 0 || move.Slot >= 8 {
			return "That defense tile is unavailable."
		}
		if g.Board&(1<<move.Slot) == 0 {
			g.PlayerDamage[move.Member] += 10
			reply = "Snowflake! You lost 10% of your maximum HP."
		} else {
			g.Defenses++
			reply = "Safe defense! The Grinch lost HP."
			if g.Defenses == g.defensePlayerCount() {
				g.Phase, g.HP, g.Board = GrinchAttack, 4500, move.NextBoard
			}
		}
	case GrinchAttack:
		if move.Slot < 0 || move.Slot >= 8 {
			return "That attack tile is unavailable."
		}
		damage := 500
		if g.Board&(1<<move.Slot) != 0 {
			damage = 525
		}
		g.HP = max(500, g.HP-damage)
		reply = "Attack hit!"
		if g.HP == 500 {
			g.Phase = GrinchFinalAttack
		}
	case GrinchFinalAttack:
		if move.Slot != 0 {
			return "Use the final attack button."
		}
		g.HP = max(0, g.HP-100)
		reply = "Final attack hit!"
		if g.HP == 0 {
			g.Phase, g.EndedAt = GrinchWon, move.At
		}
	default:
		return "This part of the fight has ended."
	}
	g.Accepted[move.RequestID] = true
	g.Dirty = true
	return reply
}
