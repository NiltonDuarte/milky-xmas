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
func TreeStatus(tree core.Tree, now time.Time, mention func(string) string) string {
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
	if core.TreeEventOver(now) {
		text.WriteString("\n🎁 The event has ended. This is the tree's final state.")
	} else {
		text.WriteString("\nIt needs 100% water to stay alive and loses 1% every hour. Use /water to help.")
	}
	if top := tree.TopWaterers(3); len(top) > 0 {
		text.WriteString("\n**Top waterers**")
		for i, waterer := range top {
			fmt.Fprintf(&text, "\n%d. %s — %d", i+1, mention(waterer.Member), waterer.Count)
		}
	}
	return text.String()
}

// Watered is the message after a successful watering.
func Watered(member string, tree core.Tree, now time.Time, mention func(string) string) string {
	return fmt.Sprintf("💧 %s watered the tree! +%d%%\n%s", mention(member), core.TreeWaterPoints, TreeStatus(tree, now, mention))
}

// TreeRejection explains why a watering did not happen. It is empty for unexpected errors.
func TreeRejection(err error) string {
	var cooldown core.TreeCooldownError
	switch {
	case errors.Is(err, core.ErrTreeFull):
		return "🎄 The tree is already full of water. Come back later!"
	case errors.As(err, &cooldown):
		minutes := int((cooldown.Remaining + time.Minute - time.Nanosecond) / time.Minute)
		var parts []string
		if hours := minutes / 60; hours > 0 {
			parts = append(parts, units(hours, "hour"))
		}
		if rest := minutes % 60; rest > 0 {
			parts = append(parts, units(rest, "minute"))
		}
		return "💧 You already watered the tree. Try again in " + strings.Join(parts, " ") + "."
	case errors.Is(err, core.ErrTreeEventOver):
		return "🎄 The event is over, so the tree can't be watered any more."
	default:
		return ""
	}
}

// TreeEventOver is the reply to a watering after the event, with the tree's final state.
func TreeEventOver(tree core.Tree, now time.Time, mention func(string) string) string {
	return TreeRejection(core.ErrTreeEventOver) + "\n" + TreeStatus(tree, now, mention)
}
