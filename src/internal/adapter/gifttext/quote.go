package gifttext

import (
	"strings"

	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

// Quote renders a Christmas quote as a quoted block with its source underneath.
func Quote(quote core.Quote) string {
	return "🎄\n> " + strings.ReplaceAll(quote.Text, "\n", "\n> ") + "\n-# " + quote.Author + ", *" + quote.Work + "*"
}
