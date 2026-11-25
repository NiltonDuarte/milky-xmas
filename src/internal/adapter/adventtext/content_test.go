package adventtext

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func TestEveryDoorHasContentThatFitsADiscordEmbed(t *testing.T) {
	seen := make(map[string]bool)
	for item := range core.AdventDays {
		text := Content(item)
		// Door heading plus the content must fit an embed description of 4096 characters.
		if strings.Contains(text, "could not be found") || text == "" || utf8.RuneCountInString(text) > 4000 || seen[text] {
			t.Errorf("content %d is missing, repeated, or too long: %d characters", item+1, utf8.RuneCountInString(text))
		}
		seen[text] = true
	}
	if !strings.Contains(Content(0), "Chicken Salpicão") {
		t.Errorf("first content = %q", Content(0))
	}
}
