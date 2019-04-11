package db

import (
	"time"

	"github.com/rs/zerolog/log"

	models "github.com/grvlle/qbot/model"
	"github.com/pkg/errors"

	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql" //Dialect
)

// Database: docker exec -it mysql1 mysql -uroot -p
type Database struct {
	*gorm.DB
}

type LastTenQuestions struct {
	ID       []uint
	Question []string
}

type LastTenAnswers struct {
	ID         []uint
	Answer     []string
	QuestionID []int
}

// InitializeDB sets up the mySQL connection
func InitializeDB() *Database {
	db, err := gorm.Open("mysql", "root:qbot@/qbot?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		log.Fatal().Msgf("Failed to connect to Database. Reason: %v\n", err)
	}
	log.Info().Msg("Successfully connected to qBot Database.")
	db.DB().SetConnMaxLifetime(time.Second * 100)
	db.DB().SetMaxIdleConns(50)
	db.DB().SetMaxOpenConns(200)
	db.DropTableIfExists(models.User{}, models.Question{}, models.Answer{}) // Temp
	if err := db.AutoMigrate(models.User{}, models.Question{}, models.Answer{}).Error; err != nil {
		log.Fatal().Msgf("Unable to migrate database. \nReason: %v", err)
	}
	log.Info().Msg("Database migration successful.")
	return &Database{db}
}

func (db *Database) CreateNewDBRecord(record interface{}) error {
	if db.NewRecord(record) != true {
		log.Warn().Msg("The value's primary key is not blank")
	}
	if err := db.Create(record).Error; err != nil {
		log.Warn().Msg("Unable to create new Database record")
		return err
	}
	log.Printf("A new Database Record were successfully added.")
	return nil
}

func (db *Database) UpdateUserTableWithQuestion(user *models.User, q *models.Question) error {
	return errors.Wrap(db.Model(&user).Find(user).Association("Questions").Append(q).Error, "Unable to update the User table with question asked")
}

func (db *Database) UpdateUserTableWithAnswer(user *models.User, a *models.Answer) error {
	return errors.Wrap(db.Model(&user).Find(user).Association("Answers").Append(a).Error, "Unable to update the User table with answer provided")
}

func (db *Database) UpdateQuestionTableWithAnswer(q *models.Question, a *models.Answer) error {
	return errors.Wrap(db.Model(&q).First(&q, a.QuestionID).Association("Answers").Append(a).Error, "Unable to update the Question table with answer provided")
}

// UpdateUsers func cross references the Users posting against the Users added
// to the DB. If a new User is detected, UpdateUsers will update the Users
// table with a new record of the poster
func (db *Database) UpdateUsers(user *models.User) {
	db.CreateNewDBRecord(user)
	if err := db.First(&user, user.ID).Error; err != nil {
		log.Warn().Msgf("Failed to add record %v to table %v.\nReason: %v", user, &user, err)
	}
}

// UserExistInDB func queries the DB for existing users prior to adding new ones.
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

// LastTenQuestions func will query the database for the last ten questions stored
// and return a populated struct of type LastTenQuestions. This function is called
// in reply.go
func (db *Database) LastTenQuestions(ltq *LastTenQuestions) *LastTenQuestions {
	tenQuestions, _ := db.Model(&models.Question{}).Order("created_at DESC").Last(&[]models.Question{}).Limit(10).Rows()
	for tenQuestions.Next() {
		q := new(models.Question)
		err := db.ScanRows(tenQuestions, q)
		if err != nil {
			log.Error().Msgf("Unable to parse SQL query into a crunchable dataformat. \nReason: %v", err)
		}
		ltq.ID = append(ltq.ID, q.ID)
		ltq.Question = append(ltq.Question, q.Question)
	}
	return ltq
}

// LastTenAnswers func will query the database for the last ten questions stored
// and return a populated struct of type LastTenAnswers. This function is called
// in reply.go
func (db *Database) LastTenAnswers(lta *LastTenAnswers) *LastTenAnswers {
	tenAnswers, _ := db.Model(&[]*models.Question{}).Related(&models.Answer{}, "Answers").Order("created_at DESC").Last(&[]models.Answer{}).Limit(10).Rows()
	for tenAnswers.Next() {
		a := new(models.Answer)
		err := db.ScanRows(tenAnswers, a)
		if err != nil {
			log.Error().Msgf("Unable to parse SQL query into a crunchable dataformat. \nReason: %v", err)
		}
		lta.ID = append(lta.ID, a.ID)
		lta.Answer = append(lta.Answer, a.Answer)
		lta.QuestionID = append(lta.QuestionID, a.QuestionID)
	}
	return lta
}
