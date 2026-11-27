package web

import (
	_ "embed"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

const fixtureServerID = "local"
const fixtureUserID = "fixture"

//go:embed chat.html
var chatPage string

func NewHandler(service app.Service, grinchStore app.GrinchStore) http.Handler {
	grinch := newGrinchFixture(time.Now, grinchStore, service.Tuning().Grinch)
	service.AuraEncounter(grinch.encounter)
	return newHandler(service, grinch, app.NewActivityDrops(service))
}

func newHandler(service app.Service, grinch grinchFixture, activity ...*app.ActivityDrops) http.Handler {
	drops := app.NewActivityDrops(service)
	if len(activity) == 1 && activity[0] != nil {
		drops = activity[0]
	}
	tree := service.RewardTree(app.NewTree(&app.MemoryTreeStore{}, time.Now).WithRules(service.Tuning().Tree))
	quotes := app.NewQuoteFeed(service)
	mux := http.NewServeMux()
	mux.HandleFunc("GET /{$}", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		io.WriteString(w, chatPage)
	})
	mux.HandleFunc("POST /chat", func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, 4096)
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Could not read the command.", http.StatusBadRequest)
			return
		}

		text := r.PostForm.Get("command")
		fields := strings.Fields(text)
		if len(fields) > 0 {
			args := fields[1:]
			switch fields[0] {
			case "/message":
				replyActivityMessage(w, r, drops, quotes, args)
				return
			case "/drop":
				replyActivityClaim(w, r, drops, args)
				return
			case "/throw":
				replyThrownSnowball(w, r, service, args)
				return
			case "/" + app.LookCommand:
				if got := strings.Join(args, "-"); got != app.StockingSubcommand {
					http.Error(w, "Use /look in-your-stocking.", http.StatusBadRequest)
					return
				}
				replyInventory(w, r, service)
				return
			case "/grinch":
				grinch.reply(w, r, args)
				return
			case "/advent":
				replyAdvent(w, r, service, args)
				return
			case "/send":
				replyGiftTransfer(w, r, service, args)
				return
			case "/water":
				if len(args) != 0 {
					http.Error(w, "Use /water.", http.StatusBadRequest)
					return
				}
				replyWater(w, r, tree)
				return
			case "/" + gifttext.SpiritCommand:
				if len(args) != 0 {
					http.Error(w, "Use /spirit.", http.StatusBadRequest)
					return
				}
				// The local chat player can use the moderator tools.
				replySpirit(w, r, service)
				return
			case "/" + gifttext.StatsCommand:
				replyStats(w, r, service, tree, args)
				return
			case "/" + gifttext.GuideCommand:
				replyGuide(w, service, args)
				return
			case "/" + gifttext.HelpCommand:
				if len(args) != 1 || args[0] != gifttext.HelpSubcommand {
					http.Error(w, "Use /help me-milky.", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				// The local chat player can use the moderator tools.
				io.WriteString(w, gifttext.Help(gifttext.HelpAccess{Moderator: true}))
				return
			case "/" + gifttext.OpenCommand:
				replyOpenGift(w, r, service, args)
				return
			case "/give":
				replyGiveGift(w, r, service, args)
				return
			case "/aura":
				replyAura(w, r, service, args)
				return
			case "/daily":
				if len(args) == 1 && args[0] == "claim" {
					replyDailyClaim(w, r, service)
					return
				}
			}
		}

		http.Error(w, "Unknown command. Try /daily claim, /advent view, /grinch view, /send gift, /open gift, /aura check, /look in-your-stocking, /water, /spirit, /stats, /guide, or /help me-milky.", http.StatusBadRequest)
	})
	return http.NewCrossOriginProtection().Handler(mux)
}

func replyActivityMessage(w http.ResponseWriter, r *http.Request, drops *app.ActivityDrops, quotes *app.QuoteFeed, args []string) {
	if len(args) == 0 {
		http.Error(w, "Use /message <text>.", http.StatusBadRequest)
		return
	}
	var replies []string
	if drop := drops.Observe(fixtureServerID, fixtureServerID); drop != nil {
		headline := "Watch out! Something is falling down!"
		if drop.Kind == core.GiftRainDrop {
			headline = "🎁 It's raining gifts! 🎁 Hallelujah"
		}
		replies = append(replies, fmt.Sprintf("%s\n🎁 🎁 🎁 🎁\n🎁 🎁 🎁 🎁\n/drop collect %s", headline, drop.ID))
	}
	// The local chat has one channel, with quotes always on.
	quote, err := quotes.Observe(r.Context(), fixtureServerID, fixtureServerID)
	if err != nil {
		http.Error(w, "Could not draw a quote. Try again.", http.StatusInternalServerError)
		return
	}
	if quote != nil {
		replies = append(replies, gifttext.Quote(*quote))
	}
	if len(replies) == 0 {
		replies = []string{"Message recorded."}
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, strings.Join(replies, "\n\n"))
}

func replyActivityClaim(w http.ResponseWriter, r *http.Request, drops *app.ActivityDrops, args []string) {
	if len(args) != 2 || (args[0] != "collect" && args[0] != "expire") {
		http.Error(w, "Use /drop collect <id> or /drop expire <id>.", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	mention := func(userID string) string { return userID }
	if args[0] == "expire" {
		// The local chat has no clock, so this ends a drop as if its time ran out.
		summary, ok := drops.Close(args[1])
		if !ok {
			io.WriteString(w, gifttext.DropRejection(app.ErrDropUnavailable))
			return
		}
		io.WriteString(w, gifttext.DropExpired(summary.Claims, mention))
		return
	}
	award, err := drops.Claim(r.Context(), app.GiftDropClaim{Member: app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}, DropID: args[1]})
	if err != nil {
		if reply := gifttext.DropRejection(err); reply != "" {
			io.WriteString(w, reply)
			return
		}
		http.Error(w, "Could not collect this gift. Try again.", http.StatusInternalServerError)
		return
	}
	io.WriteString(w, gifttext.DropCollected(award))
	if award.Finished != nil {
		io.WriteString(w, "\n"+gifttext.DropFinished(award.Finished.Claims, mention))
	}
}

func replyDailyClaim(w http.ResponseWriter, r *http.Request, service app.Service) {
	claim, err := service.ClaimGift(r.Context(), app.Member{ServerID: fixtureServerID, UserID: fixtureUserID})
	if err != nil {
		http.Error(w, "Could not record your gift claim. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.Claim(claim))
}

func replyInventory(w http.ResponseWriter, r *http.Request, service app.Service) {
	member := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	items, err := service.ViewInventory(r.Context(), member)
	if err != nil {
		http.Error(w, gifttext.InventoryUnavailable, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.Inventory(items, false))
}

func replySpirit(w http.ResponseWriter, r *http.Request, service app.Service) {
	spirit, err := service.ChristmasSpirit(r.Context(), fixtureServerID)
	if err != nil {
		http.Error(w, gifttext.SpiritUnavailable, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, strings.Join(gifttext.SpiritBlocks(spirit), "\n"))
}

// replyStats stands in for the Discord drop-downs: "/stats server" or "/stats <leaderboard>".
func replyStats(w http.ResponseWriter, r *http.Request, service app.Service, tree *app.Tree, args []string) {
	choice := gifttext.StatsServerView
	if len(args) == 1 {
		choice = args[0]
	}
	if _, ok := gifttext.Leaderboard(choice, core.EventStats{}, nil); len(args) > 1 || (!ok && choice != gifttext.StatsServerView) {
		http.Error(w, "Use /stats server, bad, good, throwers, gifters, or plants.", http.StatusBadRequest)
		return
	}
	stats, err := service.Stats(r.Context(), fixtureServerID, tree)
	if err != nil {
		http.Error(w, gifttext.StatsUnavailable, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if choice == gifttext.StatsServerView {
		io.WriteString(w, strings.Join(gifttext.ServerStats(stats), "\n"))
		return
	}
	blocks, _ := gifttext.Leaderboard(choice, stats, func(userID string) string { return userID })
	io.WriteString(w, strings.Join(blocks, "\n"))
}

func replyGuide(w http.ResponseWriter, service app.Service, args []string) {
	topic := gifttext.GuideStart
	if len(args) == 1 {
		topic = args[0]
	}
	blocks, ok := gifttext.Guide(topic, service.Tuning())
	if len(args) > 1 || !ok {
		topics := make([]string, len(gifttext.GuideTopics))
		for i, choice := range gifttext.GuideTopics {
			topics[i] = choice.Value
		}
		http.Error(w, "Use /guide "+strings.Join(topics, ", ")+".", http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, strings.Join(blocks, "\n"))
}

func replyWater(w http.ResponseWriter, r *http.Request, tree *app.Tree) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	member := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	mention := func(userID string) string { return userID }
	updated, watering, err := tree.Water(r.Context(), member)
	if err == nil || errors.Is(err, app.ErrGoodDeedNotSaved) {
		io.WriteString(w, gifttext.Watered(member.UserID, watering, updated, tree.Rules(), tree.Now(), mention))
		return
	}
	if errors.Is(err, core.ErrTreeEventOver) {
		io.WriteString(w, gifttext.TreeEventOver(updated, tree.Rules(), tree.Now()))
		return
	}
	if reply := gifttext.TreeRejection(err); reply != "" {
		io.WriteString(w, reply)
		return
	}
	http.Error(w, gifttext.TreeUnavailable, http.StatusInternalServerError)
}
