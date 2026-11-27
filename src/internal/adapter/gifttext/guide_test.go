package gifttext

import (
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestGuideTopics(t *testing.T) {
	var all strings.Builder
	for _, topic := range GuideTopics {
		blocks, ok := Guide(topic.Value, core.DefaultTuning())
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
		"/daily claim", "/look in-your-stocking", "/send gift", "/open gift", "/aura check",
		"/throw snowball", "/advent view", "/water", "/stats", "/help me-milky",
	} {
		if !strings.Contains(all.String(), command) {
			t.Errorf("the guide does not mention %s", command)
		}
	}
	if strings.Contains(all.String(), "/give") || strings.Contains(all.String(), "/setup") {
		t.Error("the guide should not list moderator commands")
	}
	if _, ok := Guide("nothing", core.DefaultTuning()); ok {
		t.Error("an unknown topic should be refused")
	}
}

func TestGuideFollowsTuning(t *testing.T) {
	tuning := core.DefaultTuning()
	tuning.Tree.WaterPoints = 7
	tuning.Tree.WaterCooldown = 90 * time.Second
	tuning.Grinch.CounterattackPercent = 0
	tree, _ := Guide("tree", tuning)
	if text := strings.Join(tree, "\n"); !strings.Contains(text, "adds 7% of water, once every 90 seconds") {
		t.Errorf("the tree topic ignores the numbers in use:\n%s", text)
	}
	grinch, _ := Guide("grinch", tuning)
	if text := strings.Join(grinch, "\n"); strings.Contains(text, "hits you back") {
		t.Errorf("a turned off counterattack should not be mentioned:\n%s", text)
	}
}
