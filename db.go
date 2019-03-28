package main

import (
	"log"

	models "./db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

//ConnectToDB sets up the mySQL connection
func ConnectToDB() (db *gorm.DB) {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Printf("Successfully connected to qBot Database.")

	return db
}

func (qb *qBot) CreateDBTables() {
	// qb.DB.DropTableIfExists(&models.Question{})
	if qb.DB.HasTable(&models.Question{}) != true {
		qb.DB.CreateTable(&models.Question{})
		log.Printf("Setting up Database Tables.")
	}
	if qb.DB.HasTable(&models.Answer{}) != true {
		qb.DB.CreateTable(&models.Answer{})
		log.Printf("Setting up Database Tables.")
	}
	log.Printf("Database check complete. No actions needed.")
}
