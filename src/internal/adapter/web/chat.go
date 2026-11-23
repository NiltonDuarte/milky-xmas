package web

import (
	_ "embed"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/gifttext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

const fixtureServerID = "local"
const fixtureUserID = "fixture"

//go:embed chat.html
var chatPage string

func NewHandler(service app.Service) http.Handler {
	return newHandler(service, newGrinchFixture(time.Now))
}

func newHandler(service app.Service, grinch grinchFixture) http.Handler {
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
			case "/grinch":
				grinch.reply(w, r, args)
				return
			case "/advent":
				replyAdvent(w, r, service, args)
				return
			case "/send":
				replyGiftTransfer(w, r, service, args)
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
			http.Error(w, "Unknown command. Try /"+app.HelloCommand+", /daily claim, /advent view, /grinch view, /send gift, or /aura check.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, reply)
	})
	return http.NewCrossOriginProtection().Handler(mux)
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
