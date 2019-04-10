package qbot

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	db "github.com/grvlle/qbot/db"
	models "github.com/grvlle/qbot/model"
	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

// Bot contains Slack API configuration data
// And provides Websocket and DB access
type Bot struct {
	// Global Bot configuration
	Config struct {
		APIToken       string   `yaml:"apiToken"`
		JoinChannels   []string `yaml:"joinChannels"`
		GeneralChannel string   `yaml:"generalChannel"`
		Debug          bool
	}

	// Websocket connection
	Slack *slack.Client
	rtm   *slack.RTM

	// IO flow
	msgCh chan Message

	// Database Connection
	DB *db.Database
}

// LoadConfig method is ran by the RunBot method and
// will populate the Config struct in the Bot type
// with configuration variables
func (qb *Bot) LoadConfig() *Bot {
	content, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(content, &qb.Config)
	if err != nil {
		panic(err)
	}
	qb.msgCh = make(chan Message, 500)
	qb.DB = db.InitializeDB()
	return qb
}

// SetupHandlers sets up the Go Routines
func (qb *Bot) SetupHandlers() {
	go qb.EventListener()
	go qb.CommandParser()
}

// Message contains the details of a recieved Slack message.
// Constructed in the EventListener method and passed in the
// messageCh
type Message struct {
	User    string
	Channel string
	Message string
}

// EventListener listens on the websocket for incoming slack
// events, including messages that it passes to the messageCh
// channel monitored by CommandParser()
func (qb *Bot) EventListener() {
	for events := range qb.rtm.IncomingEvents {
		switch ev := events.Data.(type) {
		case *slack.MessageEvent:
			msg := new(Message)
			msg.User, msg.Channel, msg.Message = ev.User, ev.Channel, ev.Text
			qb.msgCh <- *msg
		case *slack.ConnectedEvent:
			log.Println("Infos: ", ev.Info)
			log.Printf("Connection counter: %v", ev.ConnectionCount)
		case *slack.PresenceChangeEvent:
			log.Printf("Presence Change: %v\n", ev)
		case *slack.LatencyReport:
			//fmt.Printf("Current latency: %v\n", ev.Value)
		case *slack.RTMError:
			log.Printf("Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			log.Printf("Invalid credentials")
			return
		}
	}
}

// CommandParser parses the Slack messages for Bot commands
func (qb *Bot) CommandParser() {
	for msgs := range qb.msgCh {
		message := msgs.Message                        // Message recieved
		sChannel := msgs.Channel                       // Slack Channel where message were sent
		userInfo, err := qb.rtm.GetUserInfo(msgs.User) // User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		msgSplit := []rune(message)
		outMsg := qb.rtm.NewOutgoingMessage("", sChannel)

		switch { // Checks incoming message for requested bot command
		case string(msgSplit[0:3]) == "!q " || string(msgSplit[0:3]) == "!Q ":
			var reply string
			outQuestion := string(msgSplit[3:])
			q := new(models.Question)
			q.Question, q.SlackChannel = outQuestion, sChannel

			user := &models.User{
				Questions: []*models.Question{q},
				Name:      userInfo.Profile.RealNameNormalized,
				Title:     userInfo.Profile.Title,
				Avatar:    userInfo.Profile.Image32,
				SlackUser: userInfo.ID,
			}

			if qb.DB.UserExistInDB(*user) != true {
				qb.DB.UpdateUsers(user)
			}

			// Update the user_questions (m2m) and questions table with question_id and question
			if err := qb.DB.UpdateUserTableWithQuestion(user, q); err != nil {
				reply = "Someone has already asked that question. Run *!lq* to see the last questions asked"
			} else {
				reply = fmt.Sprintf("Thank you %s for providing a question. Your question has been assigned ID: %v", user.Name, q.ID)
			}
			outMsg = qb.rtm.NewOutgoingMessage(reply, sChannel)

		case string(msgSplit[0:3]) == "!lq" || string(msgSplit[0:3]) == "!LQ":
			//qb.ListQuestions(qb.Slack, sChannel)
		case string(msgSplit[0:4]) == "!qna" || string(msgSplit[0:4]) == "!QnA":
			outMsg = qb.rtm.NewOutgoingMessage("List answer and questions", sChannel)
		case string(msgSplit[0:3]) == "!a " || string(msgSplit[0:3]) == "!A ":
			reply := "Answer provided"
			parts := strings.Fields(string(msgSplit[3:])) // Splits incoming message into slice
			questionID, err := strconv.Atoi(parts[0])     // Verifies that the first element after "!a " is an intiger (Question ID)
			if err != nil {
				log.Printf("Question ID was not provided with the question answered")
				reply = fmt.Sprintf("Please include an ID for the question you're answering\n E.g '!a 123 The answer is no!'")
			}
			outAnswer := strings.Join(parts[1:], " ")
			if len(outAnswer) != 0 {

				a := new(models.Answer)
				a.Answer, a.QuestionID, a.SlackChannel = outAnswer, questionID, sChannel
				user := &models.User{
					Answers:   []*models.Answer{a},
					Name:      userInfo.Profile.RealNameNormalized,
					Title:     userInfo.Profile.Title,
					Avatar:    userInfo.Profile.Image32,
					SlackUser: userInfo.ID,
				}
				if qb.DB.UserExistInDB(*user) != true {
					qb.DB.UpdateUsers(user)
				}
				// Update the user_answers (m2m) and answers table with the answer_id and answer
				if err := qb.DB.UpdateUserTableWithAnswer(user, a); err != nil {
					log.Println(err)
					reply = "I had problems storing your provided answer in the DB"
				}
				// Update the questions_answer (m2m) table record with the answer_id
				q := models.Question{Answers: []*models.Answer{a}}
				if err := qb.DB.UpdateQuestionTableWithAnswer(&q, a); err != nil {
					log.Println(err)
					reply = "I had problems storing your provided answer in the DB. Did you specify the Question ID correctly?"
				} else {
					reply = fmt.Sprintf("Thank you %s for providing an answer to question %v. Your answer has been assigned ID: %v", user.Name, q.ID, a.ID)
				}
			} else {
				reply = "No answer was provided, please try again"
				log.Println("No answer")
			}
			outMsg = qb.rtm.NewOutgoingMessage(reply, sChannel)
		}
		qb.rtm.SendMessage(outMsg)
	}
}

// RunBot will initiate the bot
func (qb *Bot) RunBot() {
	qb.LoadConfig()
	qb.Slack = slack.New(qb.Config.APIToken)
	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm
	qb.SetupHandlers()
	qb.rtm.ManageConnection()
}
