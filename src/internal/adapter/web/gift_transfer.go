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
		http.Error(w, "Use /send gift <recipient> <rarity> <quantity>.", http.StatusBadRequest)
		return
	}
	quantity, err := strconv.ParseInt(args[3], 10, 64)
	if err != nil {
		quantity = 0
	}
	rarity := core.GiftRarity(args[2])
	sender := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	recipient := app.GiftRecipient{Member: app.Member{ServerID: fixtureServerID, UserID: args[1]}}
	goodIncrease, err := service.SendGift(r.Context(), sender, recipient, rarity, quantity)
	if err != nil {
		if reply := gifttext.TransferRejection(err); reply != "" {
			http.Error(w, reply, http.StatusBadRequest)
			return
		}
		http.Error(w, "Could not send your gifts. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.Transfer(rarity, quantity, args[1], goodIncrease))
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
