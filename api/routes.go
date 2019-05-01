package api

import "github.com/gorilla/mux"

func SetupRoutes() *mux.Router {
	// Set up API routes
	router := mux.NewRouter()

	router.HandleFunc("/api", Index)

	router.HandleFunc("/api/users", GetUsers).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/api/users/{id}", GetUser).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/api/questions", GetQuestions).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	router.HandleFunc("/api/questions/{id}", GetQuestion).
		Methods("GET").
		HeadersRegexp("Content-Type", "application/json")

	return router
}
