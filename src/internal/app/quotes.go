package app

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// quoteRolls bounds the random number that picks a quote; it is far above the number of quotes.
const quoteRolls = 1 << 20

type quoteChannel struct {
	activity core.QuoteActivity
	left     []int
	last     int
}

// QuoteStore remembers when each channel last had a quote, so a restart keeps the interval.
type QuoteStore interface {
	// LastQuote returns the zero time for a channel that never had a quote.
	LastQuote(ctx context.Context, server, channel string) (time.Time, error)
	SaveLastQuote(ctx context.Context, server, channel string, at time.Time) error
}

// QuoteFeed counts each channel's messages and hands out Christmas quotes. The counts live in
// memory, so a restart starts them over.
type QuoteFeed struct {
	service  Service
	store    QuoteStore // nil keeps the last quote times in memory only
	mu       sync.Mutex
	channels map[dropChannel]quoteChannel
}

func NewQuoteFeed(service Service) *QuoteFeed {
	store, _ := service.gifts.(QuoteStore)
	return &QuoteFeed{service: service, store: store, channels: make(map[dropChannel]quoteChannel)}
}

// Observe records one eligible human message and returns the quote it brings, if any. A quote is
// only handed out after its time was saved.
func (feed *QuoteFeed) Observe(ctx context.Context, server, channel string) (*core.Quote, error) {
	if feed == nil || server == "" || channel == "" {
		return nil, nil
	}
	feed.mu.Lock()
	defer feed.mu.Unlock()

	key := dropChannel{server: server, channel: channel}
	state, known := feed.channels[key]
	if !known {
		state.last = -1
		if feed.store != nil {
			last, err := feed.store.LastQuote(ctx, server, channel)
			if err != nil {
				return nil, fmt.Errorf("load last quote time: %w", err)
			}
			state.activity.LastQuote = last
		}
	}
	now := feed.service.now().UTC()
	activity, due := feed.service.tuning.Quotes.Observe(state.activity, now)
	if !due {
		state.activity = activity
		feed.channels[key] = state
		return nil, nil
	}
	if feed.store != nil {
		if err := feed.store.SaveLastQuote(ctx, server, channel, now); err != nil {
			return nil, fmt.Errorf("save last quote time: %w", err)
		}
	}
	state.activity = activity
	quotes := core.ChristmasQuotes()
	state.last, state.left = core.DrawQuote(state.left, state.last, len(quotes), feed.service.random(quoteRolls))
	feed.channels[key] = state
	return &quotes[state.last], nil
}
