package api

import (
	"encoding/json"
	db "github.com/grvlle/qbot/db"
	models "github.com/grvlle/qbot/model"
	"net/http"
)

var DB *db.Database

func GetUsers(w http.ResponseWriter, r *http.Request) {
	var users []models.User
	tx := DB.Begin()
	tx.Find(&users)
	json.NewEncoder(w).Encode(&users)
}
