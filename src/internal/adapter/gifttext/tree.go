package gifttext

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const TreeUnavailable = "Could not water the tree. Try again."

// treeText describes the tree: the water level and the tree's state.
func treeText(tree core.Tree, rules core.TreeRules, now time.Time) string {
	filled := tree.Level / 10
	var text strings.Builder
	fmt.Fprintf(&text, "## 🎄 The Christmas tree\n%s%s  **%d%%** 💧\n",
		strings.Repeat("🟦", filled), strings.Repeat("⬜", 10-filled), tree.Level)
	text.WriteString([]string{
		"🥀 The tree is withered. It needs a lot of water!",
		"🍂 The tree is wilting.",
		"🌱 The tree is recovering.",
		"🌲 The tree is lush. Almost there!",
		"🎄 The tree is full and alive!",
	}[core.TreeStage(tree.Level)])
	if rules.EventOver(now) {
		text.WriteString("\n🎁 The event has ended. This is the tree's final state.")
	} else {
		fmt.Fprintf(&text, "\nIt needs 100%% water to stay alive and loses 1%% every %s. Use /water to help.", every(rules.DecayInterval))
	}
	return text.String()
}

// WateredBlocks is the message after a successful watering. The good deeds are left out when the
// player's aura did not move.
func WateredBlocks(member string, watering app.Watering, tree core.Tree, rules core.TreeRules, now time.Time, mention func(string) string) []string {
	first := fmt.Sprintf("💧 %s watered the tree! +%d%%", mention(member), watering.Added)
	switch {
	case watering.ListChange > 0:
		first += fmt.Sprintf(" (😇 Nice list +%d%%)", watering.ListChange)
	case watering.ListChange < 0:
		first += fmt.Sprintf(" (😈 Naughty list %d%%)", watering.ListChange)
	}
	if watering.GoodDeeds > 0 {
		first += fmt.Sprintf("\n😇 Good deeds +%d", watering.GoodDeeds)
	}
	return []string{first, treeText(tree, rules, now)}
}

// Watered is WateredBlocks as one text.
func Watered(member string, watering app.Watering, tree core.Tree, rules core.TreeRules, now time.Time, mention func(string) string) string {
	return strings.Join(WateredBlocks(member, watering, tree, rules, now, mention), "\n")
}

// TreeRejection explains why a watering did not happen. It is empty for unexpected errors.
func TreeRejection(err error) string {
	var cooldown core.TreeCooldownError
	switch {
	case errors.Is(err, core.ErrTreeFull):
		return "🎄 The tree is fully watered. Try again later!"
	case errors.As(err, &cooldown):
		return "💧 You already watered the tree. Try again in " + wait(cooldown.Remaining) + "."
	case errors.Is(err, core.ErrTreeEventOver):
		return "🎄 The event is over, so the tree can't be watered any more."
	default:
		return ""
	}
}

// TreeEventOverBlocks is the reply to a watering after the event, with the tree's final state.
func TreeEventOverBlocks(tree core.Tree, rules core.TreeRules, now time.Time) []string {
	return []string{TreeRejection(core.ErrTreeEventOver), treeText(tree, rules, now)}
}

// TreeEventOver is TreeEventOverBlocks as one text.
func TreeEventOver(tree core.Tree, rules core.TreeRules, now time.Time) string {
	return strings.Join(TreeEventOverBlocks(tree, rules, now), "\n")
}

// wait writes a duration as hours and minutes, rounded up to the minute.
func wait(d time.Duration) string {
	minutes := int((d + time.Minute - time.Nanosecond) / time.Minute)
	var parts []string
	if hours := minutes / 60; hours > 0 {
		parts = append(parts, units(hours, "hour"))
	}
	if rest := minutes % 60; rest > 0 {
		parts = append(parts, units(rest, "minute"))
	}
	return strings.Join(parts, " ")
}

// every writes an interval for "every ...": "hour" for one hour, "30 minutes" for half of one.
func every(d time.Duration) string {
	switch text := wait(d); text {
	case "1 hour":
		return "hour"
	case "1 minute":
		return "minute"
	default:
		return text
	}
}
