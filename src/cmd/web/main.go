package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"time"

	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/giftfixture"
	"github.com/NiltonDuarte/milky-xmas/src/internal/adapter/web"
	"github.com/NiltonDuarte/milky-xmas/src/internal/app"
)

func main() {
	store := &giftfixture.Store{}
	server := &http.Server{
		Addr:         "127.0.0.1:8080",
		Handler:      web.NewHandler(app.NewService(store, time.Now, rand.IntN), store),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}
	log.Printf("Local test chat at http://%s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
