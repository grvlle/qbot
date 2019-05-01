package main

import (
	api "github.com/grvlle/qbot/api"
	database "github.com/grvlle/qbot/db"
	qb "github.com/grvlle/qbot/qbot"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var qbot qb.QBot

func init() {
	InitializeLogger()
}

func main() {
	// Initialize database connection and share across modules
	db := database.InitializeDB()
	api.DB = db
	qbot.DB = db
	defer db.Close()

	// Run Slack Bot
	go qbot.RunBot()

	// Setup API Routes
	router := api.SetupRoutes()

	// Initialize API Webserver
	srv := &http.Server{
		Handler: router,
		Addr:    "127.0.0.1:8000",
		// Enforces timeouts for webserver
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}
	log.Fatal().Err(srv.ListenAndServe())

}
