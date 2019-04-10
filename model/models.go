package model

import (
	"github.com/jinzhu/gorm"

	_ "github.com/jinzhu/gorm/dialects/mysql" //mysql
)

/*User table in the Database */
type User struct {
	gorm.Model
	Questions []*Question `gorm:"many2many:user_questions;"`
	Answers   []*Answer   `gorm:"many2many:user_answers;"`
	Name      string
	Title     string
	Avatar    string
	SlackUser string `gorm:"type:varchar(10);unique"`
}

/*Question table in the Database */
type Question struct {
	gorm.Model
	Question     string    `gorm:"type:varchar(300);unique"`
	Answers      []*Answer `gorm:"many2many:question_answers;"`
	SlackChannel string
}

/*Answer table in the Database */
type Answer struct {
	gorm.Model
	Answer       string `gorm:"type:varchar(300);unique"`
	QuestionID   int    `gorm:"index"`
	SlackChannel string
}
