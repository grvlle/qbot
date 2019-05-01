package api

import (
	"encoding/json"
	db "github.com/grvlle/qbot/db"
	//models "github.com/grvlle/qbot/model"
	"github.com/gorilla/mux"
	"github.com/rs/zerolog/log"
	"net/http"
	"strconv"
)

var DB *db.Database

func GetQuestions(w http.ResponseWriter, r *http.Request) {
	questions, err := DB.QueryQuestions()
	if err != nil {
		log.Error().Err(err)
	}
	json.NewEncoder(w).Encode(&questions)
}

func GetQuestion(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"]) // Capture ID and convert to int
	users, err := DB.QueryAnsweredQuestionsByID(id)
	if err != nil {
		log.Error().Err(err)
	}
	json.NewEncoder(w).Encode(&users)
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	users, err := DB.QueryUsers()
	if err != nil {
		log.Error().Err(err)
	}
	json.NewEncoder(w).Encode(&users)
}

func GetUser(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(mux.Vars(r)["id"]) // Capture ID and convert to int
	users, err := DB.QueryUsersByID(id)
	if err != nil {
		log.Error().Err(err)
	}
	json.NewEncoder(w).Encode(&users)
}

// func GetUsers(w http.ResponseWriter, r *http.Request) {
// 	var users []models.Question
// 	tx := DB.Begin()
// 	tx.Find(&users)
// 	json.NewEncoder(w).Encode(&users)
// }
