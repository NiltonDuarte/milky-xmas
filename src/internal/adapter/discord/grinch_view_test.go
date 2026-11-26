package discord

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/grinchtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
	"github.com/bwmarrin/discordgo"
)

// componentCount counts the components Discord counts toward the limit of 40 in one message.
func componentCount(t *testing.T, view []discordgo.MessageComponent) int {
	t.Helper()
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	var tree any
	if err := json.Unmarshal(data, &tree); err != nil {
		t.Fatal(err)
	}
	var count func(any) int
	count = func(node any) int {
		total := 0
		switch node := node.(type) {
		case map[string]any:
			if _, ok := node["type"]; ok {
				total++
			}
			for _, child := range node {
				total += count(child)
			}
		case []any:
			for _, child := range node {
				total += count(child)
			}
		}
		return total
	}
	return count(tree)
}

func viewText(t *testing.T, view []discordgo.MessageComponent) string {
	t.Helper()
	data, err := json.Marshal(view)
	if err != nil {
		t.Fatal(err)
	}
	var text strings.Builder
	var walk func(any)
	walk = func(node any) {
		switch node := node.(type) {
		case map[string]any:
			if content, ok := node["content"].(string); ok {
				text.WriteString(content + "\n")
			}
			for _, child := range node {
				walk(child)
			}
		case []any:
			for _, child := range node {
				walk(child)
			}
		}
	}
	var tree any
	if err := json.Unmarshal(data, &tree); err != nil {
		t.Fatal(err)
	}
	walk(tree)
	return text.String()
}

func fightWithPlayers(phase core.GrinchPhase, players int) core.Grinch {
	fight := core.Grinch{ID: "fight", MessageID: "999", Phase: phase, HP: 4000, StartedAt: time.Now(), Board: 0b01010101,
		Joined: map[string]bool{}, Accepted: map[string]bool{}, PlayerDamage: map[string]int{}, DamageDealt: map[string]core.GrinchDamage{},
		Gifts: map[string]int{}, Snowballs: map[string]int{}}
	for i := range players {
		member := fmt.Sprintf("p%02d", i)
		fight.Joined[member] = true
		fight.DamageDealt[member] = core.GrinchDamage{Hundredths: 1000 * (players - i)} // p00 deals the most.
	}
	return fight
}

func TestGrinchViewListsWhoJoined(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	lobby := fightWithPlayers(core.GrinchLobby, 0)
	if text := viewText(t, grinchView(rules, lobby, []string{grinchImageName}, false, false)); !strings.Contains(text, "Nobody yet") {
		t.Errorf("empty lobby = %q", text)
	}
	lobby = fightWithPlayers(core.GrinchLobby, 2)
	view := grinchView(rules, lobby, []string{grinchImageName}, false, false)
	if text := viewText(t, view); !strings.Contains(text, "**Ready (2)**\n<@p00> <@p01>") || !strings.Contains(text, "The Grinch is coming!") {
		t.Errorf("lobby = %q", text)
	}
	if len(view) != 1 {
		t.Errorf("the lobby has %d top-level components, want one panel", len(view))
	}
	many := grinchJoinedList(fightWithPlayers(core.GrinchLobby, grinchJoinedLimit+5))
	if !strings.HasSuffix(many, " and 5 more") || len(many) > 1500 {
		t.Errorf("long list = %q", many)
	}
}

func TestGrinchViewShowsPlayerCardsDuringTheFight(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	fight := fightWithPlayers(core.GrinchAttack, 2)
	fight.PlayerDamage["p01"] = 100
	view := grinchView(rules, fight, []string{grinchImageName}, false, false)
	if len(view) != 2 {
		t.Fatalf("two players give %d components, want the panel and one activity box with the cards in it", len(view))
	}
	// The cards are the last items of the activity box, after the buttons.
	box := view[1].(discordgo.Container).Components
	n := len(box)
	if box[n-4].Type() != discordgo.ActionsRowComponent || box[n-3].Type() != discordgo.SeparatorComponent ||
		box[n-2].Type() != discordgo.TextDisplayComponent || box[n-1].Type() != discordgo.TextDisplayComponent {
		t.Fatalf("the box does not end with the buttons, a separator, and the two cards")
	}
	text := viewText(t, view)
	for _, want := range []string{
		"**<@p00>**\n❤️ 100% HP · ⚔️ 20.00 damage", "**<@p01>**\n💀 Out of HP · ⚔️ 10.00 damage", "👽 Grinch · 40.00/100% HP",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("view misses %q in %q", want, text)
		}
	}
	// The card with most damage comes first.
	if strings.Index(text, "<@p00>") > strings.Index(text, "<@p01>") {
		t.Errorf("cards are not ordered by damage: %q", text)
	}
}

func TestGrinchViewCapsCardsAndStaysWithinDiscordsLimit(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	for _, phase := range []core.GrinchPhase{core.GrinchSnowballs, core.GrinchDefense, core.GrinchAttack, core.GrinchFinalAttack} {
		fight := fightWithPlayers(phase, 12)
		fight.PlayerDamage["p11"], fight.PlayerDamage["p10"] = 100, 40
		fight.FinalRound = 1
		view := grinchView(rules, fight, []string{grinchImageName}, false, false)
		text := viewText(t, view)
		if got := strings.Count(text, "**<@"); got != grinchCardLimit {
			t.Errorf("%s: %d cards, want %d", phase, got, grinchCardLimit)
		}
		// p08 to p11 are the four players without a card. p11 is out of HP, so three still stand. Their damage is 4+3+2+1 = 10.
		if want := "➕ 4 more players · 3 still standing · ⚔️ 100.00 damage"; !strings.Contains(text, want) {
			t.Errorf("%s: view misses %q in %q", phase, want, text)
		}
		if count := componentCount(t, view); count > 40 {
			t.Errorf("%s: %d components, Discord allows 40", phase, count)
		}
	}
}

func TestGrinchViewShowsFinalLeaderboards(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	fight := fightWithPlayers(core.GrinchWon, 3)
	fight.EndedAt = time.Now()
	fight.Gifts["p01"], fight.Snowballs["p02"] = 5, 2
	view := grinchView(rules, fight, []string{grinchImageName}, false, false)
	text := viewText(t, view)
	for _, want := range []string{
		"The group defeated the Grinch!", "**🏆 Damage leaderboard**\n1. <@p00>", "**🎁 Gift leaderboard**\n1. <@p01> — 5",
		"**❄️ Snowball leaderboard**\n1. <@p02> — 2",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("final view misses %q in %q", want, text)
		}
	}
	if strings.Contains(text, "<@p00> — 0") || strings.Contains(text, "<@p01> — 0 snowballs") || strings.Contains(text, "<@p00> — 0 snowballs") {
		t.Errorf("players without points are listed: %q", text)
	}
	if len(view) != 1 || strings.Contains(text, "more players") || strings.Contains(text, "HP ·") {
		t.Errorf("the final message must show leaderboards instead of cards: %d components, %q", len(view), text)
	}

	big := fightWithPlayers(core.GrinchEscaped, 25)
	big.EndedAt = time.Now()
	attached := viewText(t, grinchView(rules, big, []string{grinchImageName}, true, false))
	data, _ := json.Marshal(grinchView(rules, big, []string{grinchImageName}, true, false))
	if !strings.Contains(attached, "Full standings are attached.") || !strings.Contains(string(data), "attachment://"+grinchStandingsFile) {
		t.Errorf("25 players without the file: %q", attached)
	}
	if strings.Count(attached, "<@p") != grinchStandingsLimit { // Only the damage board has scores, and it lists ten.
		t.Errorf("the damage board must list ten players and the others none: %q", attached)
	}
	denied := viewText(t, grinchView(rules, big, []string{grinchImageName}, false, true))
	if !strings.Contains(denied, "Full standings require permission to attach files.") {
		t.Errorf("view without permission = %q", denied)
	}
	if text := viewText(t, grinchView(rules, core.Grinch{ID: "g", Phase: core.GrinchCancelled}, []string{grinchImageName}, false, false)); !strings.Contains(text, "cancelled") {
		t.Errorf("cancelled view = %q", text)
	}
}

func TestGrinchFinalMessageWithoutPermissionToAttachFiles(t *testing.T) {
	session, err := discordgo.New("Bot fixture")
	if err != nil {
		t.Fatal(err)
	}
	var attempts []string
	session.Client.Transport = transportFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/") {
			attempts = append(attempts, "with files")
			return fakeResponse(http.StatusForbidden, `{"message":"Missing Permissions","code":50013}`), nil
		}
		attempts = append(attempts, "without files")
		var edit struct {
			Content     string
			Components  []json.RawMessage
			Attachments []json.RawMessage
		}
		if err := json.Unmarshal(body, &edit); err != nil {
			t.Fatal(err)
		}
		text := string(edit.Components[0])
		if edit.Content != "" || edit.Attachments != nil || !strings.Contains(text, "Full standings require permission to attach files.") ||
			strings.Contains(text, grinchStandingsFile) || strings.Contains(text, "maxpee.jpg") || strings.Contains(text, "grinch.jpg") {
			t.Errorf("retry without files = %s", body)
		}
		return fakeResponse(http.StatusOK, `{"id":"999"}`), nil
	})
	state := fightWithPlayers(core.GrinchEscaped, 25)
	state.EndedAt = time.Now()
	if err := (grinchDisplay{session, "400", core.DefaultTuning().Grinch}).Update(t.Context(), state); err != nil {
		t.Fatal(err)
	}
	if want := []string{"with files", "without files"}; !slices.Equal(attempts, want) {
		t.Fatalf("attempts = %q, want %q", attempts, want)
	}
}

func TestGrinchViewShowsTheBossAndPlayerAreas(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	fight := fightWithPlayers(core.GrinchDefense, 2)
	text := viewText(t, grinchView(rules, fight, []string{grinchImageName}, false, false))
	for _, want := range []string{"👽 **The Grinch**\n*The Grinch is watching…*", "⚔️ **The players**\n*Nobody has attacked yet.*"} {
		if !strings.Contains(text, want) {
			t.Errorf("empty areas miss %q in %q", want, text)
		}
	}

	fight.BossFeed = []core.GrinchEvent{{Kind: core.EventPhase, Phase: core.GrinchDefense, Roll: 0}, {Kind: core.EventMiss, Member: "p01", Amount: 10, Roll: 3}}
	fight.PlayerFeed = []core.GrinchEvent{{Kind: core.EventHit, Member: "p00", Damage: 1}, {Kind: core.EventDefend, Member: "p01", Damage: 15, Roll: 0}}
	text = viewText(t, grinchView(rules, fight, []string{grinchImageName}, false, false))
	for _, want := range []string{
		"👽 **The Grinch**\n💥 “Max, mark it down: one direct hit!”\n↳ <@p01> was hit by a snowball: -10% HP\n👽 “Fine! Let's see you dodge THIS!”\n↳ Defense: hide behind the tree!",
		"⚔️ **The players**\n🎄 <@p01> · 15.00 dmg · “Nothing to see here, just a big tree.”\n⚔️ <@p00> · 1.00 dmg",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("view misses %q in %q", want, text)
		}
	}
	// The areas belong to a fight in progress, not to registration or the final message.
	lobby := viewText(t, grinchView(rules, fightWithPlayers(core.GrinchLobby, 1), []string{grinchImageName}, false, false))
	ended := fightWithPlayers(core.GrinchWon, 1)
	ended.EndedAt = time.Now()
	final := viewText(t, grinchView(rules, ended, []string{grinchImageName}, false, false))
	for name, text := range map[string]string{"lobby": lobby, "final": final} {
		if strings.Contains(text, "The Grinch**") || strings.Contains(text, "The players**") {
			t.Errorf("%s shows the fight areas: %q", name, text)
		}
	}
}

// The message must not change size as events come in: both areas always have the same number of lines.
func TestGrinchActivityAreasKeepTheirHeight(t *testing.T) {
	lines := func(state core.Grinch) (boss, players int) {
		area := grinchActivity(state, nil, nil)
		texts := area.Components
		return strings.Count(texts[0].(discordgo.TextDisplay).Content, "\n"), strings.Count(texts[2].(discordgo.TextDisplay).Content, "\n")
	}
	empty := fightWithPlayers(core.GrinchDefense, 2)
	full := fightWithPlayers(core.GrinchDefense, 2)
	for i := range 10 {
		full.BossFeed = append(full.BossFeed, core.GrinchEvent{Kind: core.EventMiss, Member: "p01", Amount: 10, Roll: i})
		full.PlayerFeed = append(full.PlayerFeed, core.GrinchEvent{Kind: core.EventDefend, Member: "p00", Damage: 1, Roll: i})
	}
	one := fightWithPlayers(core.GrinchDefense, 2)
	one.BossFeed = []core.GrinchEvent{{Kind: core.EventGift, Member: "p00", Amount: 2}}
	one.PlayerFeed = []core.GrinchEvent{{Kind: core.EventHit, Member: "p00", Damage: 1}}
	wantBoss, wantPlayers := 2*grinchtext.BossEvents, grinchtext.PlayerEvents
	for name, state := range map[string]core.Grinch{"empty": empty, "one event": one, "more events than fit": full} {
		if boss, players := lines(state); boss != wantBoss || players != wantPlayers {
			t.Errorf("%s: the areas have %d and %d lines after their titles, want %d and %d", name, boss, players, wantBoss, wantPlayers)
		}
	}
	// The activity is a box of its own, apart from the panel with the image and buttons.
	view := grinchView(core.DefaultTuning().Grinch, full, []string{grinchImageName}, false, false)
	if len(view) != 2 || view[1].(discordgo.Container).Components[0].(discordgo.TextDisplay).Content[:len("👽 **The Grinch**")] != "👽 **The Grinch**" {
		t.Errorf("the activity box is not the second component: %d top-level components", len(view))
	}
}

func TestGrinchButtonsSitInTheActivityBoxBelowTheTextAreas(t *testing.T) {
	rules := core.DefaultTuning().Grinch
	for _, phase := range []core.GrinchPhase{core.GrinchSnowballs, core.GrinchDefense, core.GrinchAttack, core.GrinchFinalAttack} {
		fight := fightWithPlayers(phase, 3)
		fight.FinalRound = 1
		view := grinchView(rules, fight, []string{grinchImageName}, false, false)
		hasRow := func(component discordgo.MessageComponent) bool {
			data, _ := json.Marshal(component)
			return strings.Contains(string(data), `"custom_id":"grinch:fight:`+string(phase)+`:`)
		}
		if hasRow(view[0]) {
			t.Errorf("%s: the panel with the image has the buttons", phase)
		}
		box := view[1].(discordgo.Container).Components
		if len(box) < 4 || box[0].Type() != discordgo.TextDisplayComponent || box[1].Type() != discordgo.SeparatorComponent ||
			box[2].Type() != discordgo.TextDisplayComponent || box[3].Type() != discordgo.ActionsRowComponent {
			t.Errorf("%s: the activity box is not the Grinch's text, a separator, the players' text, and then the buttons", phase)
		}
		for _, card := range view[2:] {
			if hasRow(card) {
				t.Errorf("%s: a player card has the buttons", phase)
			}
		}
		if data, _ := json.Marshal(view); strings.Contains(string(data), "Join") {
			t.Errorf("%s: the combat screen has a Join button", phase)
		}
	}
	// Registration keeps its Start button in the panel, and has no activity box.
	lobby := grinchView(rules, fightWithPlayers(core.GrinchLobby, 1), []string{grinchImageName}, false, false)
	data, _ := json.Marshal(lobby)
	if len(lobby) != 1 || !strings.Contains(string(data), `"label":"Start"`) {
		t.Errorf("registration = %s, want one panel with the Start button", data)
	}
}
