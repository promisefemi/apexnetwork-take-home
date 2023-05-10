package main

import (
	"github.com/boltdb/bolt"
	"github.com/go-chi/chi"
	"github.com/promisefemi/apexnetwork-take-home/handler"
	"log"
	"net/http"
)

func main() {
	port := ":9000"
	db, err := bolt.Open("my.db", 0600, nil)
	if err != nil {
		log.Fatalln(err)
	}
	router := chi.NewMux()
	pageHandler := handler.NewPageHandler(db)

	router.Post("/register", pageHandler.Register)
	router.Post("/fund-wallet", pageHandler.FundWallet)
	router.Post("/get-wallet-balance", pageHandler.GetWalletBalance)
	router.Post("/roll-dice", pageHandler.Roll)
	router.Post("/end-game", pageHandler.Roll)
	router.Post("/start-game", pageHandler.Roll)

	if err := http.ListenAndServe(port, router); err != nil {
		log.Fatalln(err)
	}
}
