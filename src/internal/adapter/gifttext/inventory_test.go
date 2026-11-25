package gifttext

import (
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestInventoryShowsOneSectionPerKind(t *testing.T) {
	items := core.Inventory{Normal: 12, Rare: 1, Snowballs: 42, YellowSnowballs: 3, Coal: 8}
	want := []string{
		"### 🎁 Gifts\n```\nNormal  12\nRare     1\nEpic     0\n```",
		"### ❄️ Snowballs\n```\n❄️ Normal  42\n🟡 Yellow   3\n```",
		"### 🌑 Coal\n```\n8\n```",
	}
	if got := InventorySections(items, false); !slices.Equal(got, want) {
		t.Errorf("sections = %q, want %q", got, want)
	}

	// Colors must not move anything: without the escape codes the text is the same.
	colored := Inventory(items, true)
	for _, painted := range []string{"\x1b[1;37mNormal\x1b[0m", "\x1b[1;32mRare\x1b[0m", "\x1b[1;35mEpic\x1b[0m", "\x1b[1;37m8\x1b[0m"} {
		if !strings.Contains(colored, painted) {
			t.Errorf("colored inventory misses %q: %q", painted, colored)
		}
	}
	stripped := regexp.MustCompile("\x1b\\[[0-9;]*m").ReplaceAllString(colored, "")
	if got := strings.ReplaceAll(stripped, "```ansi", "```"); got != Inventory(items, false) {
		t.Errorf("colored inventory without colors =\n%s", got)
	}
}
