package core

import (
	"slices"
	"testing"
)

func TestEventStatsPointsAndRanking(t *testing.T) {
	stats := EventStats{Players: map[string]PlayerActivity{
		"alice": {NormalGiftsSent: 1, RareGiftsSent: 2, EpicGiftsSent: 3, Waterings: 4, CoalSent: 20, YellowThrown: 1},
		"bob":   {CoalSent: 2},
		"carol": {CoalSent: 2},
		"dave":  {SnowballsThrown: 9},
	}}
	alice := stats.Players["alice"]
	if good, bad := alice.GoodPoints(1), alice.BadPoints(); good != 5+20+90+4 || bad != 105 {
		t.Errorf("points = %d good, %d bad; bad deeds must not lower good ones or stop at 50", good, bad)
	}
	want := []StatEntry{{"alice", 105}, {"bob", 10}, {"carol", 10}}
	if got := stats.Ranking(PlayerActivity.BadPoints); !slices.Equal(got, want) {
		t.Errorf("ranking = %+v, want %+v", got, want)
	}
	if got := stats.Total(func(a PlayerActivity) int64 { return a.CoalSent }); got != 24 {
		t.Errorf("coal total = %d", got)
	}
	if got := (EventStats{}).Ranking(PlayerActivity.BadPoints); len(got) != 0 {
		t.Errorf("empty ranking = %+v", got)
	}
}
