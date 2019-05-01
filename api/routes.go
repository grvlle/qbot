package api

import "github.com/gorilla/mux"

func SetupRoutes() *mux.Router {
	// Set up API routes
	router := mux.NewRouter()

	router.HandleFunc("/users", GetUsers).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/users/{id}", GetUser).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/questions", GetQuestions).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/questions/{id}", GetQuestion).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	return router
}
