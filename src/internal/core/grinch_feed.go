package core

// GrinchFeedSize is how many recent events each side of the fight keeps for display.
const GrinchFeedSize = 5

// GrinchEventKind says what happened in a fight event.
type GrinchEventKind string

const (
	// Player side: who is attacking and the damage it did.
	EventHit    GrinchEventKind = "hit"
	EventDefend GrinchEventKind = "defend"
	EventRevive GrinchEventKind = "revive"
	// Boss side: what the Grinch did to the players and what they took back from him.
	EventPhase    GrinchEventKind = "phase"
	EventMiss     GrinchEventKind = "miss"
	EventGift     GrinchEventKind = "gift"
	EventSnowball GrinchEventKind = "snowball"
)

// GrinchEvent is one line of the fight's recent activity. Roll is a random number the adapter
// uses to pick one of the Grinch's lines, or the line of a player hiding behind the tree, so the
// core stays free of randomness.
type GrinchEvent struct {
	Kind   GrinchEventKind
	Member string
	// Damage is the HP percent a hit or defense took from the Grinch.
	Damage float64
	// Amount is the HP percent a miss cost the player or a revival gave back, or the gift points or
	// Snowballs a player took back.
	Amount int
	// Phase is the phase the fight entered, for EventPhase.
	Phase GrinchPhase
	Roll  int
}

func addEvent(feed []GrinchEvent, event GrinchEvent) []GrinchEvent {
	feed = append(feed, event)
	if len(feed) > GrinchFeedSize {
		feed = append([]GrinchEvent(nil), feed[len(feed)-GrinchFeedSize:]...)
	}
	return feed
}

// playerEvent and bossEvent record recent activity for the two text areas of the fight message.
func (g *Grinch) playerEvent(event GrinchEvent) { g.PlayerFeed = addEvent(g.PlayerFeed, event) }
func (g *Grinch) bossEvent(event GrinchEvent)   { g.BossFeed = addEvent(g.BossFeed, event) }
