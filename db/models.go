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
	Question     string
	Answered     bool
	Answers      []Answer `gorm:"many2many:question_answers"`
	AnswerID     uint
	SlackChannel string
}

/*Answer type will be used to store answers
(to questions asked by users) in the Database */
type Answer struct {
	gorm.Model
	User         string
	Answer       string
	QuestionID   int
	SlackChannel string
}
