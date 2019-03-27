package main

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

func ConnectToDB() (db *gorm.DB) {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Printf("Successfully connected to qBot Database.")

	return db
}
