package db

import (
	"time"

	_ "github.com/jinzhu/gorm/dialects/mysql" //mysql dialect
)

/*User table in the Database */
type User struct {
	ID        int `gorm:"primary_key"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Questions []Question `gorm:"many2many:user_questions;"`
	Answers   []Answer   `gorm:"many2many:user_answers;"`
	Name      string
	Title     string
	Avatar    string
	SlackUser string `gorm:"type:varchar(10);unique"`
}

/*Question table in the Database */
type Question struct {
	ID           int `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       int      `gorm:"index"`
	Question     string   `gorm:"type:varchar(100);unique"`
	Answers      []Answer `gorm:"many2many:question_answers;"`
	SlackChannel string
}

/*Answer table in the Database */
type Answer struct {
	ID           int `gorm:"primary_key"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	UserID       int    `gorm:"index"`
	Answer       string `gorm:"type:varchar(100);unique"`
	QuestionID   int    `gorm:"index"`
	SlackChannel string
}
