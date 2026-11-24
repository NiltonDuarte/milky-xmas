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
	return newHandler(service, newGrinchFixture(time.Now, grinchStore, service.Tuning().Grinch), app.NewActivityDrops(service))
}

func newHandler(service app.Service, grinch grinchFixture, activity ...*app.ActivityDrops) http.Handler {
	drops := app.NewActivityDrops(service)
	if len(activity) == 1 && activity[0] != nil {
		drops = activity[0]
	}
	tree := app.NewTree(&app.MemoryTreeStore{}, time.Now)
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
				replyActivityMessage(w, r, drops, args)
				return
			case "/drop":
				replyActivityClaim(w, r, drops, args)
				return
			case "/thrown":
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
			case "/" + gifttext.HelpCommand:
				if len(args) != 0 {
					http.Error(w, "Use /helpmemilky.", http.StatusBadRequest)
					return
				}
				w.Header().Set("Content-Type", "text/plain; charset=utf-8")
				// The local chat player can use the moderator tools.
				io.WriteString(w, gifttext.Help(gifttext.HelpAccess{Moderator: true}))
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

		command, ok := strings.CutPrefix(text, "/")
		if !ok {
			command = ""
		}
		reply, err := service.Reply(command)
		if err != nil {
			http.Error(w, "Unknown command. Try /"+app.HelloCommand+", /daily claim, /advent view, /grinch view, /send gift, /aura check, /look in-your-stocking, /water, or /helpmemilky.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, reply)
	})
	return http.NewCrossOriginProtection().Handler(mux)
}

func replyActivityMessage(w http.ResponseWriter, r *http.Request, drops *app.ActivityDrops, args []string) {
	if len(args) == 0 {
		http.Error(w, "Use /message <text>.", http.StatusBadRequest)
		return
	}
	drop := drops.Observe(fixtureServerID, fixtureServerID)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if drop == nil {
		io.WriteString(w, "Message recorded.")
		return
	}
	if drop.Kind == "rain" {
		fmt.Fprintf(w, "🎁 It's raining gifts! 🎁 Hallelujah\n🎁 🎁 🎁 🎁\n🎁 🎁 🎁 🎁\n/drop collect %s", drop.ID)
		return
	}
	fmt.Fprintf(w, "Watch out! Something is falling down!\n🎁 🎁 🎁 🎁\n🎁 🎁 🎁 🎁\n/drop collect %s", drop.ID)
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
		io.WriteString(w, "\n"+gifttext.DropExpired(award.Finished.Claims, mention))
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
	io.WriteString(w, gifttext.Inventory(items))
}

func replyWater(w http.ResponseWriter, r *http.Request, tree *app.Tree) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	member := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	mention := func(userID string) string { return userID }
	updated, err := tree.Water(r.Context(), member)
	if err == nil {
		io.WriteString(w, gifttext.Watered(member.UserID, updated, tree.Now(), mention))
		return
	}
	if errors.Is(err, core.ErrTreeEventOver) {
		io.WriteString(w, gifttext.TreeEventOver(updated, tree.Now(), mention))
		return
	}
	if reply := gifttext.TreeRejection(err); reply != "" {
		io.WriteString(w, reply)
		return
	}
	http.Error(w, gifttext.TreeUnavailable, http.StatusInternalServerError)
}
