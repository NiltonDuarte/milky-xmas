package web

import (
	_ "embed"
	"io"
	"net/http"
	"strings"

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

		command, ok := strings.CutPrefix(r.PostForm.Get("command"), "/")
		if !ok {
			command = ""
		}
		reply, err := service.Reply(command)
		if err != nil {
			http.Error(w, "Unknown command. Try /"+app.HelloCommand+".", http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, reply)
	})
	return http.NewCrossOriginProtection().Handler(mux)
}
