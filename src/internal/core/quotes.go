package core

import (
	"slices"
	"time"
)

// QuoteRules say how much channel activity brings a Christmas quote.
type QuoteRules struct {
	// EveryMessages is how many player messages a channel needs for a quote.
	EveryMessages int
	// MinInterval is the shortest time between two quotes in one channel.
	MinInterval time.Duration
}

// DefaultQuoteRules returns the numbers used without a tuning file.
func DefaultQuoteRules() QuoteRules {
	return QuoteRules{EveryMessages: 3, MinInterval: 3 * time.Hour}
}

// QuoteActivity is what a channel did since its last quote.
type QuoteActivity struct {
	Messages  int
	LastQuote time.Time
}

// Observe counts one message and reports whether it brings a quote.
func (rules QuoteRules) Observe(activity QuoteActivity, now time.Time) (QuoteActivity, bool) {
	activity.Messages++
	tooSoon := !activity.LastQuote.IsZero() && now.Sub(activity.LastQuote) < rules.MinInterval
	if activity.Messages < rules.EveryMessages || tooSoon {
		return activity, false
	}
	return QuoteActivity{LastQuote: now}, true
}

// DrawQuote picks the next quote among those left in the round, starting a new round of total
// quotes when none are left. The last quote shown is never picked twice in a row. The roll is
// any non-negative uniform number.
func DrawQuote(left []int, last, total, roll int) (picked int, rest []int) {
	rest = slices.Clone(left)
	if len(rest) == 0 {
		for i := range total {
			rest = append(rest, i)
		}
	}
	candidates := slices.DeleteFunc(slices.Clone(rest), func(i int) bool { return i == last })
	if len(candidates) == 0 {
		candidates = rest
	}
	picked = candidates[roll%len(candidates)]
	return picked, slices.DeleteFunc(rest, func(i int) bool { return i == picked })
}

// Quote is a Christmas quote or verse. Text may have several lines.
type Quote struct {
	Text   string
	Author string
	Work   string
}

// ChristmasQuotes are the quotes the bot shows. They are all in the public domain; keep it so,
// because the repository is public.
func ChristmasQuotes() []Quote {
	const (
		dickens, carol = "Charles Dickens", "A Christmas Carol"
		moore, visit   = "Clement Clarke Moore", "A Visit from St. Nicholas"
		longfellow     = "Henry Wadsworth Longfellow"
		bells          = "Christmas Bells"
		rossetti       = "Christina Rossetti"
		midwinter      = "In the Bleak Midwinter"
	)
	return []Quote{
		{"I will honour Christmas in my heart, and try to keep it all the year.", dickens, carol},
		{"God bless us, every one!", dickens, carol},
		{"I have always thought of Christmas time … as a good time;\na kind, forgiving, charitable, pleasant time.", dickens, carol},
		{"It is good to be children sometimes, and never better than at Christmas.", dickens, carol},
		{"Happy, happy Christmas, that can win us back to the delusions of our childish days.", dickens, "The Pickwick Papers"},
		{"'Twas the night before Christmas, when all through the house\nNot a creature was stirring, not even a mouse;\nThe stockings were hung by the chimney with care,\nIn hopes that St. Nicholas soon would be there.", moore, visit},
		{"The children were nestled all snug in their beds,\nWhile visions of sugar-plums danced in their heads.", moore, visit},
		{"But I heard him exclaim, ere he drove out of sight,\n\"Happy Christmas to all, and to all a good night!\"", moore, visit},
		{"I heard the bells on Christmas Day\nTheir old, familiar carols play,\nAnd wild and sweet\nThe words repeat\nOf peace on earth, good-will to men!", longfellow, bells},
		{"Then pealed the bells more loud and deep:\n\"God is not dead, nor doth He sleep;\nThe Wrong shall fail,\nThe Right prevail,\nWith peace on earth, good-will to men.\"", longfellow, bells},
		{"In the bleak mid-winter\nFrosty wind made moan,\nEarth stood hard as iron,\nWater like a stone.", rossetti, midwinter},
		{"What can I give Him,\nPoor as I am?\nIf I were a shepherd\nI would bring a lamb;\nIf I were a Wise Man\nI would do my part;\nYet what I can I give Him:\nGive my heart.", rossetti, midwinter},
		{"Heap on more wood! the wind is chill;\nBut let it whistle as it will,\nWe'll keep our Christmas merry still.", "Walter Scott", "Marmion"},
		{"The time draws near the birth of Christ:\nThe moon is hid; the night is still;\nThe Christmas bells from hill to hill\nAnswer each other in the mist.", "Alfred Tennyson", "In Memoriam"},
		{"\"Christmas won't be Christmas without any presents,\"\ngrumbled Jo, lying on the rug.", "Louisa May Alcott", "Little Women"},
		{"Yes, Virginia, there is a Santa Claus. He exists as certainly as love and generosity and devotion exist.", "Francis P. Church", "The Sun, 1897"},
		{"The magi, as you know, were wise men, wonderfully wise men, who brought gifts to the Babe in the manger.\nThey invented the art of giving Christmas presents.", "O. Henry", "The Gift of the Magi"},
	}
}
