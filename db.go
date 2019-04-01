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

	db.AutoMigrate(models.User{}, models.Question{}, models.Answer{})

	return db
}

//TODO: Make sure this method takes multiple arguments
func (qb *qBot) CreateDBTables(tables ...interface{}) error {
	for _, table := range tables {
		if qb.DB.HasTable(table) != true {
			if err := qb.DB.CreateTable(table).Error; err != nil {
				log.Fatalf("Unable to create Database tables.\nReason: %v", err)
				qb.DB.Close()
				return err
			}
			log.Printf("Setting up Database Tables.")
		}
	}
	log.Printf("Database check complete. No actions needed.")
	return nil
}

func (qb *qBot) CreateNewDBRecord(DBtable interface{}) error {
	if qb.DB.NewRecord(DBtable) != true {
		log.Printf("The value's primary key is not blank")
	}
	if err := qb.DB.Create(DBtable).Error; err != nil {
		log.Printf("Unable to create new Database record")
		return err
	}
	log.Printf("A new Database Record were successfully added.")
	return nil
}

func (qb *qBot) UserExistInDB(newUserRecord models.User) bool {
	var count int64
	if err := qb.DB.Where("slack_user = ?", newUserRecord.SlackUser).First(&newUserRecord).Count(&count); err != nil { //Count DB entries matching the Slack User ID
		if count == 0 { //Avoid duplicate User entries in the DB.
			return false
		}
	}
	return true
}
