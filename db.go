package main

import (
	"log"

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

//TODO: Make sure this method takes multiple arguments
func (qb *qBot) CreateDBTables(DBtable interface{}) error {
	if qb.DB.HasTable(DBtable) != true {
		if err := qb.DB.CreateTable(DBtable).Error; err != nil {
			log.Fatalf("Unable to create Database tables.\nReason: %v", err)
			qb.DB.Close()
			return err
		}
		log.Printf("Setting up Database Tables.")
	}
	qb.DB.AutoMigrate()
	log.Printf("Database check complete. No actions needed.")
	return nil
}

func (qb *qBot) CreateNewDBRecord(DBtable interface{}) error {
	if qb.DB.NewRecord(DBtable) != true {
		log.Printf("The value's primary key is not blank")
	}
	if err := qb.DB.Create(DBtable).Error; err != nil {
		log.Fatalf("Unable to create new Database record")
		qb.DB.Close()
		return err
	}
	log.Printf("A new Database Record were successfully added.")
	return nil
}
