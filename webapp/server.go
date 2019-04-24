package main

import (api "github.com/grvlle/qbot/api")

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/resources", GetResources).Methods("GET")
	router.HandleFunc("/resources/{id}", GetResource).Methods("GET")
	router.HandleFunc("/resources", CreateResource).Methods("POST")
	router.HandleFunc("/resources/{id}", DeleteResource).Methods("DELETE")

	log.Fatal(http.ListenAndServe(":8000", router))
}
