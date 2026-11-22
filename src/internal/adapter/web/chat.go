package web

import (
	_ "embed"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

//go:embed chat.html
var chatPage string

func NewHandler(service app.Service) http.Handler {
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
		if len(fields) > 0 && fields[0] == "/advent" {
			replyAdvent(w, r, service, fields[1:])
			return
		}
		if len(fields) == 2 && fields[0] == "/daily" && fields[1] == "claim" {
			replyDailyClaim(w, r, service)
			return
		}

		command, ok := strings.CutPrefix(text, "/")
		if !ok {
			command = ""
		}
		reply, err := service.Reply(command)
		if err != nil {
			http.Error(w, "Unknown command. Try /"+app.HelloCommand+" or /daily claim.", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, reply)
	})
	return http.NewCrossOriginProtection().Handler(mux)
}

func replyDailyClaim(w http.ResponseWriter, r *http.Request, service app.Service) {
	claim, err := service.ClaimGift(r.Context(), app.Member{ServerID: "local", UserID: "fixture"})
	if err != nil {
		http.Error(w, "Could not record your gift claim. Try again.", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	if claim.Collected {
		fmt.Fprintf(w, "🎁 You collected one gift! Total gifts: %d.", claim.Total)
		return
	}
	next := claim.NextClaim.Add(time.Second - time.Nanosecond).UTC().Format("2006-01-02 15:04:05 MST")
	fmt.Fprintf(w, "Your next gift is available %s.", next)
}
