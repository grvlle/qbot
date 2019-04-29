package main

import (
	api "github.com/grvlle/qbot/api"
	database "github.com/grvlle/qbot/db"
	qb "github.com/grvlle/qbot/qbot"
	"github.com/rs/zerolog/log"
	"net/http"

	"github.com/gorilla/mux"
)

var bot qb.QBot

func init() {
	InitializeLogger()
}

func main() {
	db := database.InitializeDB()
	api.DB = db
	bot.DB = db
	defer db.Close()

	go bot.RunBot()

	router := mux.NewRouter()
	router.HandleFunc("/users", api.GetUsers).Methods("GET")
	// router.HandleFunc("/resources/{id}", GetResource).Methods("GET")
	// router.HandleFunc("/resources", CreateResource).Methods("POST")
	// router.HandleFunc("/resources/{id}", DeleteResource).Methods("DELETE")
	log.Fatal().Err(http.ListenAndServe(":8000", router))

}
