package db

import (
	"fmt"
	"io/ioutil"
	"time"

	models "github.com/grvlle/qbot/model"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" // MySQL
)

const (
	queryLimit = 5 // Limit the amount of records retrieved across all DB query functions
)

type Database struct {
	*gorm.DB
}

type dbConfig struct {
	Database struct {
		DatabaseType     string `yaml:"type"`
		Database         string `yaml:"database"`
		DatabaseUser     string `yaml:"user"`
		DatabasePassword string `yaml:"password"`
	}
}

// InitializeDB sets up the mySQL connection and migrates the database
func InitializeDB() *Database {
	config := new(dbConfig)
	configFile, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		panic(err)
	}
	credentials := fmt.Sprintf("%s:%s@/%s?charset=utf8&parseTime=True&loc=Local", config.Database.DatabaseUser, config.Database.DatabasePassword, config.Database.Database)
	dialect := config.Database.DatabaseType

	db, err := gorm.Open(dialect, credentials)
	if err != nil {
		log.Fatal().Msgf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Info().Msg("Successfully connected to qBot Database.")

	db.DB().SetConnMaxLifetime(time.Second * 100)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(200)

	// db.DropTableIfExists(models.User{}, models.Question{}, models.Answer{}) // Temp
	// db.DropTable("user_questions", "question_answers", "user_answers")      // Temp
	if err := db.AutoMigrate(models.User{}, models.Question{}, models.Answer{}).Error; err != nil {
		log.Fatal().Msgf("Unable to migrate database. \nReason: %v", err)
	}
	log.Info().Msg("Database migration successful.")
	return &Database{db}
}

/* CREATE FUNCTIONS */

// CreateNewDBRecord checks if record exists. If not, proceeds to create a new one
// TODO: Rewrite using FirstorCreate method insted
func (db *Database) CreateNewDBRecord(record interface{}) error {
	if !db.NewRecord(record) {
		log.Warn().Msg("The value's primary key is not blank")
	}
	if err := db.Create(record).Error; err != nil {
		log.Warn().Msg("Unable to create new Database record")
		return err
	}
	log.Info().Msg("A new Database Record were successfully added.")
	return nil
}

// UpdateUserTableWithQuestion updates the User m2m relationship with questions asked
func (db *Database) UpdateUserTableWithQuestion(user *models.User, q *models.Question) error {
	return errors.Wrap(db.Model(&user).Find(user).Association("Questions").Append(q).Error, "Unable to update the User table with question asked")
}

// UpdateUserTableWithAnswer updates the User m2m relationship with answers provided
func (db *Database) UpdateUserTableWithAnswer(user *models.User, a *models.Answer) error {
	return errors.Wrap(db.Model(&user).Find(user).Association("Answers").Append(a).Error, "Unable to update the User table with answer provided")
}

// UpdateQuestionTableWithAnswer updates the Question tables m2m relationship with answers provided
func (db *Database) UpdateQuestionTableWithAnswer(q *models.Question, a *models.Answer) error {
	return errors.Wrap(db.Model(&q).First(&q, a.QuestionID).Association("Answers").Append(a).Error, "Unable to update the Question table with answer provided")
}

/* READ FUNCTIONS */

// UserExistInDB func queries the DB for existing users prior to adding new ones.
// TODO: Rewrite this function to use Firstorcreate method instead.
func (db *Database) UserExistInDB(newUserRecord models.User) bool {
	var count int64
	// Count DB entries matching the Slack User ID
	if err := db.Where("slack_user = ?", newUserRecord.SlackUser).First(&newUserRecord).Count(&count); err != nil {
		if count == 0 { // Avoid duplicate User entries in the DB.
			return false
		}
	}
	return true
}

// QueryQuestions func will query the database for the last ten questions stored
// and return a DB object of Questions
func (db *Database) QueryQuestions() ([]models.Question, error) {
	var lq []models.Question
	return lq, errors.Wrap(db.Limit(queryLimit).Model(&models.Question{}).Order("id DESC").Find(&lq).Error, "Unable to query Questions table for last questions")
}

// QueryAnsweredQuestionsByID queries the Questions table by ID and its m2m Answer relationship
// for questions and answers. It returns a db object containing the information. This is parsed
// and buffered using the ParseQueryAndCacheContent func.
func (db *Database) QueryAnsweredQuestionsByID(questionID int) ([]models.Question, error) {
	var qna []models.Question
	return qna, errors.Wrap(db.Limit(queryLimit).Model(&models.Question{}).Related(&[]models.Answer{}, "Answers").Preload("Answers").Order("id DESC").First(&qna, questionID).Error, "Unable to query database")
}

// QueryAnsweredQuestions queries the most recent Questions table and its m2m Answer relationship
// for questions and answers. It returns a db object containing the information. This is parsed
// and buffered using the ParseQueryAndCacheContent func.
func (db *Database) QueryAnsweredQuestions() ([]models.Question, error) {
	var qna []models.Question
	return qna, errors.Wrap(db.Limit(queryLimit).Model(&models.Question{}).Related(&[]models.Answer{}, "Answers").Preload("Answers").Order("id DESC").Find(&qna).Error, "Unable to query database")
}

// QueryQuestionsAskedByUserID : TODO
func (db *Database) QueryQuestionsAskedByUserID(slackID string) ([]models.User, error) {
	var user []models.User
	return user, errors.Wrap(db.Model(&models.User{}).Related(&[]models.Question{}, "Questions").Preload("Questions").Where("slack_user = ?", slackID).Find(&user).Error, "Unable to query user by ID")
}

// QueryUsersByID : TODO
func (db *Database) QueryUsersByID(userID int) ([]models.User, error) {
	var user []models.User
	return user, errors.Wrap(db.Model(&models.User{}).Related(&[]models.Question{}, "Questions").Related(&[]models.Answer{}, "Answers").Preload("Answers").Preload("Questions").Find(&user).Error, "Unable to query user")
}

// QueryUsers : TODO
func (db *Database) QueryUsers() ([]models.User, error) {
	var user []models.User
	return user, errors.Wrap(db.Model(&models.User{}).Related(&[]models.Question{}, "Questions").Related(&[]models.Answer{}, "Answers").Find(&user).Error, "Unable to query users")
}

/* UPDATE FUNCTIONS */

// UpdateUsers func cross references the user object against the Users in
// the DB. If a new User is detected, UpdateUsers will update the Users
// table with a new record of the poster
func (db *Database) UpdateUsers(user *models.User) uint {
	db.CreateNewDBRecord(user)
	if err := db.First(&user, user.ID).Error; err != nil {
		log.Warn().Msgf("Failed to add record %v to table %v.\nReason: %v", user, &user, err)
	}
	return user.ID
}

/* DELETE FUNCTIONS */

// DeleteAnsweredQuestionsByID queries the Questions table by ID and its m2m Answer relationship
// for questions and answers. It returns a db object containing the information. This is parsed
// and buffered using the ParseQueryAndCacheContent func.
// TODO: Make sure m2m table record also gets deleted
func (db *Database) DeleteAnsweredQuestionsByID(questionID int) error {
	var q []models.Question
	return errors.Wrap(db.Model(&models.Question{}).Related(&[]models.Answer{}, "Answers").Preload("Answers").Where("id = ?", questionID).Unscoped().Delete(&q).Error, "Unable to delete user from Database")
}
