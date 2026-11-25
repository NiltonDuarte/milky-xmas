package discord

import (
	"fmt"
	"slices"
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/grinchtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

// The fight message is a Components V2 message. Discord allows 40 components in one: the largest
// board (the attack grid and Join) takes 16 with the main container, and each player card takes two.
const (
	grinchCardLimit      = 8  // Players with a card of their own, most damage first.
	grinchStandingsLimit = 10 // Players listed on each final leaderboard.
	grinchJoinedLimit    = 40 // Mentions in the registration list, so the text stays well under 4000 characters.

	grinchStandingsFile = "grinch-leaderboard.txt"

	colorLobby    = 0x2ECC71
	colorFight    = 0xE74C3C
	colorWon      = 0xF1C40F
	colorEnded    = 0x95A5A6
	colorActivity = 0x1E1F22 // The darkest tone, so the activity box stands out from the panel above it.
)

// grinchView is the whole fight message. Images are the attachment names shown at the top.
// Standings that do not fit are written to the standings file; denied says the bot may not attach it.
func grinchView(rules core.GrinchRules, state core.Grinch, images []string, attachFile, denied bool) []discordgo.MessageComponent {
	content, rows := grinchMessage(rules, state)
	panel := []discordgo.MessageComponent{discordgo.TextDisplay{Content: content}}
	if len(images) > 0 {
		panel = append([]discordgo.MessageComponent{grinchGallery(images)}, panel...)
	}
	accent := colorFight
	switch {
	case state.Phase == core.GrinchCancelled:
		return []discordgo.MessageComponent{grinchContainer(colorEnded, []discordgo.MessageComponent{discordgo.TextDisplay{Content: content}})}
	case state.Phase == core.GrinchLobby:
		accent = colorLobby
		panel = append(panel, discordgo.Separator{}, discordgo.TextDisplay{Content: grinchJoinedList(state)})
	case !state.EndedAt.IsZero():
		accent = colorEnded
		if state.Phase == core.GrinchWon {
			accent = colorWon
		}
		sections, overflow := grinchStandings(rules, state)
		panel = append(panel, discordgo.Separator{})
		for _, section := range sections {
			panel = append(panel, discordgo.TextDisplay{Content: section})
		}
		switch {
		case overflow && attachFile:
			panel = append(panel, discordgo.TextDisplay{Content: "Full standings are attached."},
				discordgo.FileComponent{File: discordgo.UnfurledMediaItem{URL: "attachment://" + grinchStandingsFile}})
		case overflow && denied:
			panel = append(panel, discordgo.TextDisplay{Content: "Full standings require permission to attach files."})
		}
	}
	active := state.EndedAt.IsZero() && state.Phase != core.GrinchLobby && state.Joined != nil
	if !active {
		panel = append(panel, rows...) // The Start button of the registration.
	}
	view := []discordgo.MessageComponent{grinchContainer(accent, panel)}
	if active {
		view = append(view, grinchActivity(state, rows, grinchCards(rules, state)))
	}
	return view
}

func mentionUser(userID string) string { return "<@" + userID + ">" }

// blankLine keeps a line of the text areas at its height when there is nothing to show, so the
// message does not change size as events come in. Discord drops plain spaces, but not this one.
const blankLine = "⠀"

// fixedLines pads or cuts lines to exactly n, so the areas keep the same height.
func fixedLines(lines []string, n int) string {
	lines = lines[:min(n, len(lines))]
	for len(lines) < n {
		lines = append(lines, blankLine)
	}
	return strings.Join(lines, "\n")
}

// grinchActivity is a box of its own with two text areas of fixed height, what the Grinch says and
// did to the players and who is attacking and the damage they do, and below them the buttons of
// the phase, and below those a card for each player. Everything shares the one box.
func grinchActivity(state core.Grinch, buttons, cards []discordgo.MessageComponent) discordgo.Container {
	boss := grinchtext.BossLines(state.BossFeed, mentionUser, grinchtext.BossEvents)
	if len(boss) == 0 {
		boss = []string{"*The Grinch is watching…*"}
	}
	players := grinchtext.PlayerLines(state.PlayerFeed, mentionUser, grinchtext.PlayerEvents)
	if len(players) == 0 {
		players = []string{"*Nobody has attacked yet.*"}
	}
	box := []discordgo.MessageComponent{
		discordgo.TextDisplay{Content: "👽 **The Grinch**\n" + fixedLines(boss, 2*grinchtext.BossEvents)},
		discordgo.Separator{},
		discordgo.TextDisplay{Content: "⚔️ **The players**\n" + fixedLines(players, grinchtext.PlayerEvents)},
	}
	box = append(box, buttons...)
	if len(cards) > 0 {
		box = append(box, discordgo.Separator{})
		box = append(box, cards...)
	}
	return grinchContainer(colorActivity, box)
}

func grinchContainer(accent int, components []discordgo.MessageComponent) discordgo.Container {
	return discordgo.Container{AccentColor: &accent, Components: components}
}

func grinchGallery(images []string) discordgo.MediaGallery {
	items := make([]discordgo.MediaGalleryItem, len(images))
	for i, name := range images {
		items[i] = discordgo.MediaGalleryItem{Media: discordgo.UnfurledMediaItem{URL: "attachment://" + name}}
	}
	return discordgo.MediaGallery{Items: items}
}

// grinchJoinedList names the players who pressed Start, in a stable order.
func grinchJoinedList(state core.Grinch) string {
	members := make([]string, 0, len(state.Joined))
	for member := range state.Joined {
		members = append(members, member)
	}
	slices.Sort(members)
	if len(members) == 0 {
		return "🎅 **Ready**\nNobody yet. Press Start!"
	}
	var list strings.Builder
	for i, member := range members {
		if i == grinchJoinedLimit {
			fmt.Fprintf(&list, " and %d more", len(members)-i)
			break
		}
		if i > 0 {
			list.WriteString(" ")
		}
		list.WriteString("<@" + member + ">")
	}
	return fmt.Sprintf("🎅 **Ready (%d)**\n%s", len(members), list.String())
}

// grinchCards shows the players with the most damage and one line for the others.
func grinchCards(rules core.GrinchRules, state core.Grinch) []discordgo.MessageComponent {
	scores := state.Leaderboard(rules)
	cards := make([]discordgo.MessageComponent, 0, grinchCardLimit+1)
	for _, score := range scores[:min(grinchCardLimit, len(scores))] {
		hp := max(0, 100-state.PlayerDamage[score.Member])
		marker, status := "🟩", fmt.Sprintf("❤️ %d%% HP", hp) // The marker stands in for a colour, as a card sits in the same box.
		switch {
		case hp == 0:
			marker, status = "⬛", "💀 Out of HP"
		case hp <= 25:
			marker = "🟧"
		case hp <= 50:
			marker = "🟨"
		}
		cards = append(cards, discordgo.TextDisplay{Content: fmt.Sprintf("%s **<@%s>**\n%s · ⚔️ %.2f damage", marker, score.Member, status, score.Points)})
	}
	if len(scores) > grinchCardLimit {
		others, standing, damage := scores[grinchCardLimit:], 0, 0.0
		for _, score := range others {
			if state.PlayerDamage[score.Member] < 100 {
				standing++
			}
			damage += score.Points
		}
		cards = append(cards, discordgo.TextDisplay{Content: fmt.Sprintf("➕ %d more players · %d still standing · ⚔️ %.2f damage", len(others), standing, damage)})
	}
	return cards
}

// grinchStandings writes the final leaderboards, ten players each. The second result says that
// more players took part than fit, so the full list goes in the standings file.
func grinchStandings(rules core.GrinchRules, state core.Grinch) ([]string, bool) {
	var sections []string
	overflow := false
	for _, board := range []struct {
		title, format, unit string
		scores              []core.GrinchScore
	}{
		{"🏆 Damage leaderboard", "%.2f", "dmg", core.WithPoints(state.Leaderboard(rules))},
		{"🎁 Gift leaderboard", "%.0f", "pts", core.WithPoints(state.GiftLeaderboard())},
		{"❄️ Snowball leaderboard", "%.0f", "snowballs", core.WithPoints(state.SnowballLeaderboard())},
	} {
		if len(board.scores) == 0 {
			continue
		}
		var section strings.Builder
		section.WriteString("**" + board.title + "**")
		for i, score := range board.scores[:min(grinchStandingsLimit, len(board.scores))] {
			fmt.Fprintf(&section, "\n%d. <@%s> — "+board.format+" %s", i+1, score.Member, score.Points, board.unit)
		}
		overflow = overflow || len(board.scores) > grinchStandingsLimit
		sections = append(sections, section.String())
	}
	return sections, overflow
}

// grinchStandingsText lists every player on every leaderboard, for the attached file.
func grinchStandingsText(rules core.GrinchRules, state core.Grinch) string {
	var text strings.Builder
	for _, board := range []struct {
		title, format, unit string
		scores              []core.GrinchScore
	}{
		{"Damage leaderboard", "%.2f", "dmg", core.WithPoints(state.Leaderboard(rules))},
		{"Gift leaderboard", "%.0f", "pts", core.WithPoints(state.GiftLeaderboard())},
		{"Snowball leaderboard", "%.0f", "snowballs", core.WithPoints(state.SnowballLeaderboard())},
	} {
		if len(board.scores) == 0 {
			continue
		}
		fmt.Fprintf(&text, "%s\n", board.title)
		for i, score := range board.scores {
			fmt.Fprintf(&text, "%d. Discord user %s — "+board.format+" %s\n", i+1, score.Member, score.Points, board.unit)
		}
		text.WriteString("\n")
	}
	return text.String()
}
