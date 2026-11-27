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

// SnowflakeHit is the reply for a missed defense that cost the given share of maximum HP.
func SnowflakeHit(percent int) string {
	return fmt.Sprintf("Snowflake! You lost %d%% of your maximum HP.", percent)
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
	// Perks is what the player's list gives them. It counts when the player enters the fight.
	Perks *GrinchPerks
	// CounterRoll is a random number in [0, 100) that decides whether the Grinch hits back.
	CounterRoll int
	// Revive, when set, spends one of the fallen player's gifts of that rarity to get back up.
	Revive GiftRarity
	// Shield says the player keeps a shield, which blocks the next hit.
	Shield bool
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
	Perks          map[string]GrinchPerks // Players on a list, as they entered the fight.
	LastMove       map[string]time.Time   // Each player's last accepted action, for the cooldown.
	DownAt         time.Time              // When the last player standing fell, if one may still get up.
	Blocked        map[string]int         // Hits each player's shields stopped.
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
	if down := g.LastChanceEnds(rules); !down.IsZero() && !g.hasSurvivor() && !now.Before(down) {
		if deadline := g.Deadline(rules); deadline.IsZero() || down.Before(deadline) {
			g.end(rules, GrinchDefeat, down)
			g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchDefeat, Roll: int(down.UnixMilli())})
			return true
		}
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
	entering := !g.Joined[move.Member]
	if entering {
		delete(g.Perks, move.Member)
		if move.Perks != nil {
			if g.Perks == nil {
				g.Perks = make(map[string]GrinchPerks)
			}
			g.Perks[move.Member] = *move.Perks
		}
	}
	reply := g.join(rules, move)
	if entering && !g.Joined[move.Member] {
		delete(g.Perks, move.Member)
	}
	return reply
}

// perks is how the fight treats the player.
func (g Grinch) perks(rules GrinchRules, member string) GrinchPerks {
	if perks, ok := g.Perks[member]; ok {
		return perks
	}
	return GrinchPerks{MissHPLoss: rules.DefenseMissHPLoss, LossCoal: rules.LossCoal}
}

// boosted adds the player's damage bonus to an attack.
func (perks GrinchPerks) boosted(damage int) int { return damage * (100 + perks.DamageBonus) / 100 }

func (g *Grinch) join(rules GrinchRules, move GrinchMove) string {
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

// HitDamage is what one hit takes from the Grinch in a phase that has to remove span, whose
// damage is fixed when the hits do not grow with the players.
func (g Grinch) hitDamage(rules GrinchRules, span, fixed int) int {
	needed := max(rules.MinHits, rules.HitsPerPlayer*len(g.Joined))
	if needed <= 0 {
		return fixed
	}
	return (span + needed - 1) / needed
}

// FinalDamage is what one final attack takes from the Grinch, before any list bonus.
func (g Grinch) FinalDamage(rules GrinchRules) int {
	return g.hitDamage(rules, rules.AttackBossHP, rules.FinalHitDamage)
}

// ReviveHP is the HP, in percent, a gift of the rarity brings a fallen player back with.
func (rules GrinchRules) ReviveHP(rarity GiftRarity) int {
	switch rarity {
	case NormalGift:
		return rules.ReviveNormalHP
	case RareGift:
		return rules.ReviveRareHP
	case EpicGift:
		return rules.ReviveEpicHP
	default:
		return 0
	}
}

// CanRevive reports whether any gift can bring a fallen player back.
func (rules GrinchRules) CanRevive() bool {
	return rules.ReviveNormalHP > 0 || rules.ReviveRareHP > 0 || rules.ReviveEpicHP > 0
}

// LastChanceEnds is when the Grinch wins unless somebody gets back up. It is zero while somebody
// is standing.
func (g Grinch) LastChanceEnds(rules GrinchRules) time.Time {
	if g.DownAt.IsZero() || !g.EndedAt.IsZero() {
		return time.Time{}
	}
	return g.DownAt.Add(rules.LastChance)
}

// SomeoneDown reports whether a player in the fight is out of HP.
func (g Grinch) SomeoneDown() bool {
	for member := range g.Joined {
		if g.PlayerDamage[member] >= 100 {
			return true
		}
	}
	return false
}

// block stops a hit with the player's shield, and reports whether there was one.
func (g *Grinch) block(move GrinchMove) bool {
	if !move.Shield {
		return false
	}
	if g.Blocked == nil {
		g.Blocked = make(map[string]int)
	}
	g.Blocked[move.Member]++
	g.bossEvent(GrinchEvent{Kind: EventShield, Member: move.Member, Roll: move.Roll})
	return true
}

// hurt takes HP from a player and records it for the boss' text area. When nobody is left standing
// the Grinch wins, after the last chance to get back up if gifts can do that.
func (g *Grinch) hurt(rules GrinchRules, move GrinchMove, percent int) {
	g.PlayerDamage[move.Member] += percent
	g.bossEvent(GrinchEvent{Kind: EventMiss, Member: move.Member, Amount: percent, Roll: move.Roll})
	if g.PlayerDamage[move.Member] < 100 || g.hasSurvivor() {
		return
	}
	if rules.CanRevive() && rules.LastChance > 0 {
		g.DownAt = move.At
		return
	}
	g.end(rules, GrinchDefeat, move.At)
	g.bossEvent(GrinchEvent{Kind: EventPhase, Phase: GrinchDefeat, Roll: move.Roll})
}

// counterattack may let the Grinch hit back at a player who attacked, and returns the words that
// tell them so.
func (g *Grinch) counterattack(rules GrinchRules, move GrinchMove) string {
	if rules.CounterattackPercent <= 0 || move.CounterRoll < 100-rules.CounterattackPercent || !g.EndedAt.IsZero() {
		return ""
	}
	if g.block(move) {
		return " Your shield blocked the Grinch's counterattack."
	}
	percent := g.perks(rules, move.Member).MissHPLoss
	g.hurt(rules, move, percent)
	return fmt.Sprintf(" The Grinch hit you back: -%d%% HP.", percent)
}

// revive gets a fallen player back up with the gift they chose.
func (g *Grinch) revive(rules GrinchRules, move GrinchMove) string {
	switch g.Phase {
	case GrinchSnowballs, GrinchDefense, GrinchAttack, GrinchFinalAttack:
	default:
		return "This part of the fight has ended."
	}
	if g.PlayerDamage[move.Member] < 100 {
		return "You are still standing. A gift only brings you back after you fall."
	}
	hp := rules.ReviveHP(move.Revive)
	if hp <= 0 {
		return "That gift cannot bring you back."
	}
	g.PlayerDamage[move.Member] = 100 - hp
	g.bossEvent(GrinchEvent{Kind: EventRevive, Member: move.Member, Amount: hp})
	g.accept(move)
	return fmt.Sprintf("You are back with %d%% HP!", hp)
}

// accept counts an action and starts the player's wait for the next one.
func (g *Grinch) accept(move GrinchMove) {
	if g.LastMove == nil {
		g.LastMove = make(map[string]time.Time)
	}
	g.LastMove[move.Member] = move.At
	g.Accepted[move.RequestID], g.Dirty = true, true
	if g.hasSurvivor() {
		g.DownAt = time.Time{}
	}
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
	if move.Revive != "" {
		return g.revive(rules, move)
	}
	if g.Phase != move.Phase || move.At.Before(g.StartedAt) {
		return "This part of the fight has ended."
	}
	if g.PlayerDamage[move.Member] >= 100 {
		return "You have no HP left in this fight."
	}
	if last, ok := g.LastMove[move.Member]; ok && move.At.Sub(last) < rules.ClickCooldown {
		return "Slow down! Wait a moment before your next move."
	}
	var reply string
	damage := GrinchDamage{}
	perks := g.perks(rules, move.Member)
	switch g.Phase {
	case GrinchSnowballs:
		if move.Slot != 0 {
			return "The snowball phase has ended."
		}
		applied := min(perks.boosted(g.hitDamage(rules, GrinchBossHP-rules.SnowballsBossHP, rules.SnowballDamage)), g.HP-rules.SnowballsBossHP)
		g.HP -= applied
		damage.Hundredths = applied
		reply = fmt.Sprintf("Snowball hit for %s HP.", FormatHundredths(applied))
		if g.HP == rules.SnowballsBossHP {
			g.DefensePlayers = rules.DefensesNeeded(len(g.Joined))
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
			if g.block(move) {
				reply = "Your shield blocked the snowflake!"
			} else {
				reply = SnowflakeHit(perks.MissHPLoss)
				g.hurt(rules, move, perks.MissHPLoss)
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
		hit := g.hitDamage(rules, rules.DefenseBossHP-rules.AttackBossHP, rules.AttackDamage)
		if g.Board&(1<<move.Slot) != 0 {
			hit = hit * rules.AttackHeavyDamage / max(1, rules.AttackDamage)
		}
		damageApplied := min(g.HP-rules.AttackBossHP, perks.boosted(hit))
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
			applied := min(g.HP, perks.boosted(g.FinalDamage(rules)))
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
	g.accept(move)
	// Only an attack provokes the Grinch: a defense, a gift, or a Snowball does not.
	if damage.Hundredths != 0 {
		reply += g.counterattack(rules, move)
	}
	return reply
}
