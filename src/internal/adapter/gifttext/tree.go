package gifttext

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const TreeUnavailable = "Could not water the tree. Try again."

// TreeStatus describes the tree. Mention formats a player's user ID.
func TreeStatus(tree core.Tree, mention func(string) string) string {
	filled := tree.Level / 10
	var text strings.Builder
	fmt.Fprintf(&text, "## 🎄 The Christmas tree\n```\n💧 Water level: %d%%  [%s%s]\n```\n", tree.Level,
		strings.Repeat("█", filled), strings.Repeat("░", 10-filled))
	text.WriteString([]string{
		"🥀 The tree is withered. It needs a lot of water!",
		"🍂 The tree is wilting.",
		"🌱 The tree is recovering.",
		"🌲 The tree is lush. Almost there!",
		"🎄 The tree is full and alive!",
	}[core.TreeStage(tree.Level)])
	fmt.Fprintf(&text, "\nIt needs 100%% water to stay alive and loses %d%% every hour. Use /water to help.", 1)
	if top := tree.TopWaterers(3); len(top) > 0 {
		text.WriteString("\n**Top waterers**")
		for i, waterer := range top {
			fmt.Fprintf(&text, "\n%d. %s — %d", i+1, mention(waterer.Member), waterer.Count)
		}
	}
	return text.String()
}

// Watered is the message after a successful watering.
func Watered(member string, tree core.Tree, mention func(string) string) string {
	return fmt.Sprintf("💧 %s watered the tree! +%d%%\n%s", mention(member), core.TreeWaterPoints, TreeStatus(tree, mention))
}

// TreeRejection explains why a watering did not happen. It is empty for unexpected errors.
func TreeRejection(err error) string {
	var cooldown core.TreeCooldownError
	switch {
	case errors.Is(err, core.ErrTreeFull):
		return "🎄 The tree is already full of water. Come back later!"
	case errors.As(err, &cooldown):
		minutes := int((cooldown.Remaining + time.Minute - time.Nanosecond) / time.Minute)
		unit := "minutes"
		if minutes == 1 {
			unit = "minute"
		}
		return fmt.Sprintf("💧 You watered the tree a moment ago. Try again in %d %s.", minutes, unit)
	default:
		return ""
	}
}
