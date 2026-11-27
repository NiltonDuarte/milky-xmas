package grinchtext

import (
	"fmt"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// The Grinch's lines are short so each fits on one line of the message. An event's random roll
// picks one of them, so the same fight can be replayed with the same lines.
var (
	linesStart = []string{
		"Ho ho NO! This is MY Christmas now!",
		"Who woke me up? I was hating joy.",
		"Snowballs? Call that a welcome gift?",
		"Max, fetch my grumpy hat. Visitors!",
		"I came for the cookies. Staying for chaos.",
	}
	linesDefense = []string{
		"Fine! Let's see you dodge THIS!",
		"My heart grew two sizes… of pure rage.",
		"Duck! Or don't. I'm easy.",
		"Snowflakes, assemble! They're tiny, but many.",
	}
	linesAttack = []string{
		"Something, something… ouch!",
		"You're on my last nerve. I only have one.",
		"Max, a little help? Max? MAX!",
		"Okay, that one actually stung.",
	}
	linesFinal = []string{
		"Not my stolen presents!",
		"I regret nothing. Except that last part.",
		"Fine, fine! Take your socks and go!",
	}
	linesDefeat = []string{
		"Ha! Nobody left to ruin my Christmas!",
		"Naughty list: everyone. Case closed.",
		"Max, put the kettle on. We won.",
	}
	// What the Grinch says when his snowball hits a player in the defense phase.
	linesMiss = []string{
		"Bullseye! Right in the jingle bells.",
		"Splat! Fresh from my snowball cannon.",
		"Too slow, Whoville! Cold enough?",
		"Max, mark it down: one direct hit!",
		"Ha! Snow business like show business.",
	}
	// What a player says when they hide behind the tree in the defense phase.
	linesHide = []string{
		"Nothing to see here, just a big tree.",
		"I live here now. Me and the tree.",
		"This tree has great snow insurance.",
		"Branch protection enabled!",
		"Tree-mendous cover!",
	}
	linesGift = []string{
		"Hey, put that gift back!",
		"My favourite present to steal!",
		"Rude. That wrapping paper was pricey.",
		"Give it back… oh, you can't. Fine.",
	}
	linesSnowball = []string{
		"Ugh, a snowball. My one weakness.",
		"Cold! Cold! Why is it always cold?",
		"Max chewed that one. You're welcome.",
	}
)

// How many events each text area shows. The Grinch's take two lines each.
const (
	BossEvents   = 3
	PlayerEvents = 5
)

func pick(lines []string, roll int) string {
	if roll < 0 {
		roll = -roll
	}
	return lines[roll%len(lines)]
}

func phaseLine(phase core.GrinchPhase, roll int) string {
	switch phase {
	case core.GrinchDefense:
		return pick(linesDefense, roll)
	case core.GrinchAttack:
		return pick(linesAttack, roll)
	case core.GrinchFinalAttack:
		return pick(linesFinal, roll)
	case core.GrinchDefeat:
		return pick(linesDefeat, roll)
	default:
		return pick(linesStart, roll)
	}
}

func phaseNote(phase core.GrinchPhase) string {
	switch phase {
	case core.GrinchDefense:
		return "Defense: hide behind the tree!"
	case core.GrinchAttack:
		return "Attack phase!"
	case core.GrinchFinalAttack:
		return "Final attack!"
	case core.GrinchDefeat:
		return "All players are down."
	default:
		return "The fight begins!"
	}
}

func count(n int, singular string) string {
	if n == 1 {
		return fmt.Sprintf("1 %s", singular)
	}
	return fmt.Sprintf("%d %ss", n, singular)
}

// BossLines writes the latest events of the Grinch, newest first, two lines each: what he says
// and what it did. At most limit events are shown. Mention formats a player's user ID.
func BossLines(feed []core.GrinchEvent, mention func(string) string, limit int) []string {
	lines := make([]string, 0, 2*min(limit, len(feed)))
	for i := len(feed) - 1; i >= 0 && len(lines) < 2*limit; i-- {
		event := feed[i]
		switch event.Kind {
		case core.EventPhase:
			lines = append(lines, fmt.Sprintf("👽 “%s”", phaseLine(event.Phase, event.Roll)), "↳ "+phaseNote(event.Phase))
		case core.EventMiss:
			lines = append(lines, fmt.Sprintf("💥 “%s”", pick(linesMiss, event.Roll)),
				fmt.Sprintf("↳ %s was hit by a snowball: -%d%% HP", mention(event.Member), event.Amount))
		case core.EventShield:
			lines = append(lines, "👽 “What is THAT?!”", fmt.Sprintf("↳ %s's shield blocked the hit 🛡️", mention(event.Member)))
		case core.EventRevive:
			lines = append(lines, "👽 “Back already?!”", fmt.Sprintf("↳ %s got back up with %d%% HP", mention(event.Member), event.Amount))
		case core.EventGift:
			lines = append(lines, fmt.Sprintf("🎁 “%s”", pick(linesGift, event.Roll)),
				fmt.Sprintf("↳ %s took back %s", mention(event.Member), count(event.Amount, "gift point")))
		case core.EventSnowball:
			lines = append(lines, fmt.Sprintf("❄️ “%s”", pick(linesSnowball, event.Roll)),
				fmt.Sprintf("↳ %s took back %s", mention(event.Member), count(event.Amount, "Snowball")))
		}
	}
	return lines
}

// PlayerLines writes who attacked and the damage it did, newest first, one line each. At
// most limit events are shown.
func PlayerLines(feed []core.GrinchEvent, mention func(string) string, limit int) []string {
	lines := make([]string, 0, min(limit, len(feed)))
	for i := len(feed) - 1; i >= 0 && len(lines) < limit; i-- {
		event := feed[i]
		switch event.Kind {
		case core.EventHit:
			lines = append(lines, fmt.Sprintf("⚔️ %s · %.2f dmg", mention(event.Member), event.Damage))
		case core.EventDefend:
			lines = append(lines, fmt.Sprintf("🎄 %s · %.2f dmg · “%s”", mention(event.Member), event.Damage, pick(linesHide, event.Roll)))
		}
	}
	return lines
}
