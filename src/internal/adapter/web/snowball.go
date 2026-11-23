package web

import (
	"io"
	"net/http"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func replyThrownSnowball(w http.ResponseWriter, r *http.Request, service app.Service, args []string) {
	if len(args) != 2 || args[0] != "snowball" {
		http.Error(w, "Use /thrown snowball <recipient>.", http.StatusBadRequest)
		return
	}
	name := args[1]
	sender := app.Member{ServerID: fixtureServerID, UserID: fixtureUserID}
	recipient := app.Recipient{Member: app.Member{ServerID: fixtureServerID, UserID: name}}
	total, err := service.ThrowSnowball(r.Context(), sender, recipient)
	if err != nil {
		if reply := gifttext.SnowballRejection(err); reply != "" {
			http.Error(w, reply, http.StatusBadRequest)
			return
		}
		http.Error(w, "Could not record your snowball throw. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, gifttext.SnowballThrown(name)+"\n"+gifttext.SnowballReceived(fixtureUserID, total))
}
