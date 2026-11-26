package core

import (
	"testing"
	"time"
)

func TestQuoteFollowsMessagesAndInterval(t *testing.T) {
	rules := QuoteRules{EveryMessages: 3, MinInterval: time.Hour}
	start := time.Date(2026, time.December, 1, 12, 0, 0, 0, time.UTC)
	var activity QuoteActivity
	for i, step := range []struct {
		at  time.Duration
		due bool
	}{
		{0, false}, {0, false}, {0, true}, // the third message brings the first quote
		{time.Minute, false}, {time.Minute, false}, {time.Minute, false}, // count reached, too soon
		{59 * time.Minute, false}, {time.Hour, true},
		{3 * time.Hour, false}, // the count starts over after a quote
	} {
		var due bool
		if activity, due = rules.Observe(activity, start.Add(step.at)); due != step.due {
			t.Fatalf("message %d at %s: due = %v, want %v", i+1, step.at, due, step.due)
		}
	}
}

func TestDrawQuoteShowsEveryQuoteBeforeRepeating(t *testing.T) {
	const total = 5
	var left []int
	last := -1
	for round := range 3 {
		seen := map[int]bool{}
		for i := range total {
			picked, rest := DrawQuote(left, last, total, round*7+i)
			if picked < 0 || picked >= total || seen[picked] || picked == last {
				t.Fatalf("round %d draw %d picked %d after %d, seen %v", round, i, picked, last, seen)
			}
			seen[picked], last, left = true, picked, rest
		}
	}
	if picked, rest := DrawQuote(nil, 0, 1, 9); picked != 0 || len(rest) != 0 {
		t.Fatalf("single quote = %d, rest %v", picked, rest)
	}
}

func TestChristmasQuotesAreComplete(t *testing.T) {
	for i, quote := range ChristmasQuotes() {
		if quote.Text == "" || quote.Author == "" || quote.Work == "" || len(quote.Text) > 1500 {
			t.Errorf("quote %d is incomplete or too long: %+v", i, quote)
		}
	}
}
