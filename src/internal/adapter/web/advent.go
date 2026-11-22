package web

import (
	"io"
	"net/http"
	"strconv"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/adventtext"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func replyAdvent(w http.ResponseWriter, r *http.Request, service app.Service, args []string) {
	member := app.Member{ServerID: "local", UserID: "fixture"}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch {
	case len(args) == 1 && args[0] == "view":
		calendar, err := service.ViewAdvent(r.Context(), member)
		if err != nil {
			http.Error(w, "Could not load your Advent calendar. Try again.", http.StatusInternalServerError)
			return
		}
		io.WriteString(w, adventtext.Calendar(calendar))
	case len(args) == 2 && args[0] == "open":
		day, err := strconv.Atoi(args[1])
		if err != nil {
			http.Error(w, "Choose an Advent door from 1 to 24.", http.StatusBadRequest)
			return
		}
		opening, err := service.OpenAdvent(r.Context(), member, day)
		if err != nil {
			http.Error(w, "Could not open the Advent door. Try again.", http.StatusInternalServerError)
			return
		}
		io.WriteString(w, adventtext.Opening(day, opening))
	default:
		http.Error(w, "Use /advent view or /advent open <day>.", http.StatusBadRequest)
	}
}
