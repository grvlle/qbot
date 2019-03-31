package db

import (
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

/*Question type will be used to store questions
asked by users in the Database */
type Question struct {
	gorm.Model
	User         string
	UserID       int `gorm:"index"`
	Question     string
	Answers      []Answer
	SlackChannel string
}

/*Answer type will be used to store answers
(to questions asked by users) in the Database */
type Answer struct {
	gorm.Model
	User         string
	UserID       int    `gorm:"index"`
	Answer       string `gorm:"type:varchar(100);unique_index"`
	QuestionID   int    `gorm:"index"`
	SlackChannel string
}

type User struct {
	gorm.Model
	Name    string
	Title   string
	Avatar  string
	SlackID string `gorm:"type:varchar(100);unique_index"`
}
