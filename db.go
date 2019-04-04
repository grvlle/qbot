package main

import (
	"log"
	"time"

	models "./db"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

//InitializeDB sets up the mySQL connection
func InitializeDB() (db *gorm.DB) {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Printf("Successfully connected to qBot Database.")

	db.DB().SetConnMaxLifetime(time.Second * 100)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(200)

	db.AutoMigrate(models.User{}, models.Question{}, models.Answer{})

	return db
}

func (qb *qBot) CreateNewDBRecord(record interface{}) error {
	if qb.DB.NewRecord(record) != true {
		log.Printf("The value's primary key is not blank")
	}
	if err := qb.DB.Create(record).Error; err != nil {
		log.Printf("Unable to create new Database record")
		return err
	}
	log.Printf("A new Database Record were successfully added.")
	return nil
}

func (qb *qBot) UserExistInDB(newUserRecord models.User) bool {
	var count int64
	//Count DB entries matching the Slack User ID
	if err := qb.DB.Where("slack_user = ?", newUserRecord.SlackUser).First(&newUserRecord).Count(&count); err != nil {
		if count == 0 { //Avoid duplicate User entries in the DB.
			return false
		}
	}
	return true
}
