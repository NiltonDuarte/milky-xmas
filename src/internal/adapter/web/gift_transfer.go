package web

import (
	"io"
	"net/http"
	"strconv"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
	"github.com/NiltonDuarte/milky-xmas/src/internal/core"
)

func replyGiftTransfer(w http.ResponseWriter, r *http.Request, service app.Service, args []string) {
	if len(args) != 4 || args[0] != "gift" {
		http.Error(w, "Use /send gift <recipient> <item> <quantity>.", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		quantity = 0
	}
	item := args[2]
	sender := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	recipient := app.Recipient{Member: app.Member{ServerID: fixtureServerID, UserID: args[1]}}
	err = service.SendItem(r.Context(), sender, recipient, item, quantity)
	if err != nil {
		if reply := gifttext.TransferRejection(err); reply != "" {
			http.Error(w, reply, http.StatusBadRequest)
			return
		}
		http.Error(w, "Could not send your gifts. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.Transfer(item, quantity, args[1]))
}

func replyAura(w http.ResponseWriter, r *http.Request, service app.Service, args []string) {
	if len(args) != 1 || args[0] != "check" {
		http.Error(w, "Use /aura check.", http.StatusBadRequest)
		return
	}
	member := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	aura, err := service.CheckAura(r.Context(), member)
	if err != nil {
		http.Error(w, "Could not load your aura. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.Aura(aura))
}

func replyGiveGift(w http.ResponseWriter, r *http.Request, service app.Service, args []string) {
	if len(args) != 4 || args[0] != "gift" {
		http.Error(w, "Use /give gift <recipient> <amount> <rarity>.", http.StatusBadRequest)
		return
	}
	amount, err := strconv.ParseInt(args[2], 10, 64)
	if err != nil {
		amount = 0
	}
	moderator := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	recipient := app.Recipient{Member: app.Member{ServerID: fixtureServerID, UserID: args[1]}}
	if err := service.GiveGift(r.Context(), moderator, recipient, core.GiftRarity(args[3]), amount); err != nil {
		if reply := gifttext.TransferRejection(err); reply != "" {
			http.Error(w, reply, http.StatusBadRequest)
			return
		}
		http.Error(w, "Could not give the gifts. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.GiftGiven(amount, args[3], args[1]))
}
