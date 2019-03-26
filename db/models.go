package db

import (
	"log"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

/*Question type will be used to store questions
asked by users in the Database */
type Question struct {
	gorm.Model
	User         string
	Question     string
	Answered     bool
	Answers      []Answer `gorm:"many2many:question_answers"`
	AnswerID     uint     `sql:"index"`
	SlackChannel string
}

type Answer struct {
	gorm.Model
	Answer string
}

func ConnectToDB() {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Printf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Printf("Successfully connected to qBot Database.")
	defer db.Close()

	db.DropTableIfExists(&Question{}, &Answer{})
	db.CreateTable(&Question{}, &Answer{})

	q := Question{
		User:     "Janne",
		Question: "Hur mar du?",
		Answered: false,
	}
	q2 := Question{
		User:     "ne",
		Question: "2Hur mar du?",
		Answered: false,
	}
	a := Answer{
		Answer: "JA!",
	}
	db.Create(&q)
	db.Create(&q2)
	db.Create(&a)
}
