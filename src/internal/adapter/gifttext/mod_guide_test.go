package gifttext

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestModGuideTopics(t *testing.T) {
	var all strings.Builder
	for _, topic := range ModGuideTopics {
		blocks, ok := ModGuide(topic.Value, core.DefaultTuning())
		if !ok || len(blocks) == 0 {
			t.Fatalf("topic %q has no text", topic.Value)
		}
		text := strings.Join(blocks, "\n")
		// The plain reply used when the box is refused takes at most 2000 characters.
		if length := utf8.RuneCountInString(text); length > 1900 {
			t.Errorf("topic %q is too long: %d characters", topic.Value, length)
		}
		if strings.Contains(text, "%!") {
			t.Errorf("topic %q has a formatting mistake:\n%s", topic.Value, text)
		}
		all.WriteString(text)
	}
	for _, command := range []string{
		"/daily claim", "/look in-your-stocking", "/send gift", "/open gift", "/aura check", "/throw snowball",
		"/advent view", "/water", "/stats", "/guide", "/help me-milky", "/give gift", "/grinch summon", "/spirit",
		"/mod-guide", "/setup view", "/setup channel",
	} {
		if !strings.Contains(all.String(), command) {
			t.Errorf("the moderator guide does not mention %s", command)
		}
	}
	if _, ok := ModGuide("nothing", core.DefaultTuning()); ok {
		t.Error("an unknown topic should be refused")
	}
}

func TestModGuideFollowsTuning(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Grinch.HitsPerPlayer, tuning.Grinch.MinHits, tuning.Grinch.SpawnChancePercent = 5, 12, 33
	fight, _ := ModGuide("fight", tuning)
	for _, want := range []string{"5 hits per player, at least 12", "1: 12 hits", "10: 50 hits"} {
		if text := strings.Join(fight, "\n"); !strings.Contains(text, want) {
			t.Errorf("fight numbers are missing %q:\n%s", want, text)
		}
	}
	grinch, _ := ModGuide("grinch", tuning)
	if text := strings.Join(grinch, "\n"); !strings.Contains(text, "33% chance") {
		t.Errorf("the spawn chance in use is missing:\n%s", text)
	}
}
