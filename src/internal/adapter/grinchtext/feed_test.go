package grinchtext

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func mention(userID string) string { return "<@" + userID + ">" }

func TestBossLinesShowWhatHeSaysAndDidNewestFirst(t *testing.T) {
	feed := []core.GrinchEvent{
		{Kind: core.EventPhase, Phase: core.GrinchSnowballs, Roll: 0},
		{Kind: core.EventMiss, Member: "1", Amount: 10, Roll: 1},
		{Kind: core.EventGift, Member: "2", Amount: 3, Roll: 2},
		{Kind: core.EventSnowball, Member: "3", Amount: 1, Roll: 0},
		{Kind: core.EventGift, Member: "4", Amount: 1, Roll: 3},
	}
	want := []string{
		"🎁 “Give it back… oh, you can't. Fine.”", "↳ <@4> took back 1 gift point",
		"❄️ “Ugh, a snowball. My one weakness.”", "↳ <@3> took back 1 Snowball",
		"🎁 “Rude. That wrapping paper was pricey.”", "↳ <@2> took back 3 gift points",
	}
	if got := BossLines(feed, mention, BossEvents); !slices.Equal(got, want) {
		t.Fatalf("boss lines =\n%q\nwant\n%q", got, want)
	}
	older := BossLines(feed[:2], mention, BossEvents)
	wantOlder := []string{
		"💥 “Splat! Fresh from my snowball cannon.”", "↳ <@1> was hit by a snowball: -10% HP",
		"👽 “Ho ho NO! This is MY Christmas now!”", "↳ The fight begins!",
	}
	if !slices.Equal(older, wantOlder) {
		t.Fatalf("boss lines =\n%q\nwant\n%q", older, wantOlder)
	}
}

func TestPlayerLinesShowWhoAttackedAndTheDamage(t *testing.T) {
	feed := []core.GrinchEvent{
		{Kind: core.EventHit, Member: "1", Damage: 1},
		{Kind: core.EventDefend, Member: "2", Damage: 4.285714, Roll: 3},
		{Kind: core.EventMiss, Member: "3"}, // Not an attack, so it is not listed.
	}
	want := []string{"🎄 <@2> · 4.29 dmg · “Branch protection enabled!”", "⚔️ <@1> · 1.00 dmg"}
	if got := PlayerLines(feed, mention, PlayerEvents); !slices.Equal(got, want) {
		t.Fatalf("player lines = %q, want %q", got, want)
	}
	if got := PlayerLines(feed, mention, 1); len(got) != 1 {
		t.Fatalf("a limit of one gave %q", got)
	}
}

func TestEveryLineFitsOnOneLineOfTheMessage(t *testing.T) {
	longest := 0
	for _, pool := range [][]string{linesStart, linesDefense, linesAttack, linesFinal, linesDefeat, linesMiss, linesHide, linesGift, linesSnowball} {
		for _, line := range pool {
			longest = max(longest, utf8.RuneCountInString(line))
		}
	}
	if longest > 48 {
		t.Errorf("the longest line has %d characters, want at most 48 so it does not wrap", longest)
	}
}

func TestTheGrinchHasDifferentLinesForEachMoment(t *testing.T) {
	for name, pool := range map[string][]string{
		"start": linesStart, "defense": linesDefense, "attack": linesAttack, "final": linesFinal, "defeat": linesDefeat,
		"miss": linesMiss, "hide": linesHide, "gift": linesGift, "snowball": linesSnowball,
	} {
		if len(pool) < 3 {
			t.Errorf("%s has %d lines, want a pool of at least three", name, len(pool))
		}
		seen := map[string]bool{}
		for _, line := range pool {
			if strings.TrimSpace(line) == "" || seen[line] {
				t.Errorf("%s has an empty or repeated line %q", name, line)
			}
			seen[line] = true
		}
	}
	// The roll picks the line, any roll is safe, and every line can be reached.
	reached := map[string]bool{}
	for roll := -3; roll < 10; roll++ {
		reached[pick(linesMiss, roll)] = true
	}
	if len(reached) != len(linesMiss) {
		t.Errorf("rolls reach %d of %d lines", len(reached), len(linesMiss))
	}
	if phaseLine(core.GrinchDefense, 1) != linesDefense[1] || phaseLine(core.GrinchDefeat, 2) != linesDefeat[2] || phaseLine(core.GrinchSnowballs, 0) != linesStart[0] {
		t.Error("a phase line comes from the wrong pool")
	}
}
