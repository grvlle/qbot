package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"strconv"
	"strings"

	"github.com/jinzhu/gorm"

	models "./db"
	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

type qBot struct {
	//Global qBot configuration
	Config struct {
		APIToken       string   `yaml:"apiToken"`
		JoinChannels   []string `yaml:"joinChannels"`
		GeneralChannel string   `yaml:"generalChannel"`
		Debug          bool
	}

	//Websocket connection
	Slack *slack.Client
	rtm   *slack.RTM

	//IO flow
	qCh   chan models.Question
	aCh   chan models.Answer
	uCh   chan models.User
	msgCh chan Message

	//Database connection
	DB *gorm.DB
}

/*LoadConfig method is ran by the RunBot method and
will populate the Config struct in the qBot type
with configuration variables*/
func (qb *qBot) LoadConfig() *qBot {

	content, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(content, &qb.Config)
	if err != nil {
		panic(err)
	}
	qb.qCh = make(chan models.Question, 500)
	qb.aCh = make(chan models.Answer, 500)
	qb.uCh = make(chan models.User)
	qb.msgCh = make(chan Message, 500)

	return qb
}

func (qb *qBot) SetupHandlers() {
	go qb.EventListener()
	go qb.CommandParser()
	go qb.AddToDatabase()
}

/*Message contains the details of a recieved
Slack message. Constructed in the EventListener
method and passed in the messageCh*/
type Message struct {
	User    string
	Channel string
	Message string
}

/* EventListener listens on the websocket for
incoming slack events, including messages that it
passes to the messageCh channel monitored by
CommandParser() */
func (qb *qBot) EventListener() {

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

/*CommandParser parses the Slack messages
for qBot commands */
func (qb *qBot) CommandParser() {

	for msgs := range qb.msgCh {
		message := msgs.Message                    //Message recieved
		sChannel := msgs.Channel                   //Slack Channel where message were sent
		user, err := qb.rtm.GetUserInfo(msgs.User) //User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		msgSplit := []rune(message)
		outMsg := qb.rtm.NewOutgoingMessage("", sChannel)

		u := new(models.User)
		u.Name, u.Title, u.Avatar, u.SlackID = user.Profile.RealNameNormalized, user.Profile.Title, user.Profile.Image32, user.ID

		switch { //Checks incoming message for requested bot command
		case string(msgSplit[0:3]) == "!q " || string(msgSplit[0:3]) == "!Q ":
			outQuestion := string(msgSplit[3:])
			q := new(models.Question)
			q.User, q.Question, q.SlackChannel = user.Profile.RealNameNormalized, outQuestion, sChannel
			outMsg = qb.rtm.NewOutgoingMessage("Question stored!", sChannel)

			qb.uCh <- *u
			qb.qCh <- *q

		case string(msgSplit[0:3]) == "!lq" || string(msgSplit[0:3]) == "!LQ":
			ListQuestions(qb.Slack, sChannel)
		case string(msgSplit[0:4]) == "!qna" || string(msgSplit[0:4]) == "!QnA":
			outMsg = qb.rtm.NewOutgoingMessage("List answer and questions", sChannel)
		case string(msgSplit[0:3]) == "!a " || string(msgSplit[0:3]) == "!A ":
			reply := "Answer provided"

			parts := strings.Fields(string(msgSplit[3:])) //Splits incoming message into slice
			questionID, err := strconv.Atoi(parts[0])     //Verifies that the first element after "!a " is an intiger (Question ID)
			if err != nil {
				log.Printf("%v failed to provide a Question ID to the question answered", user.RealName)
				reply = fmt.Sprintf("Please include an ID for the question you're answering\n E.g '!a 123 The answer is no!'")
			}
			outAnswer := parts[1]
			a := new(models.Answer)
			a.User, a.Answer, a.QuestionID, a.SlackChannel = user.Profile.RealName, outAnswer, questionID, sChannel
			outMsg = qb.rtm.NewOutgoingMessage(reply, sChannel)

			qb.uCh <- *u
			qb.aCh <- *a

		}
		qb.rtm.SendMessage(outMsg)
	}
}

/*AddToDatabase creates new records in the qBot DB*/
func (qb *qBot) AddToDatabase() {

	newQuestionRecord := models.Question{}
	newAnswerRecord := models.Answer{}
	newUserRecord := models.User{}

	for {
		select {
		case newQuestionRecord = <-qb.qCh:
			qb.CreateNewDBRecord(&newQuestionRecord)
			if err := qb.DB.First(&newQuestionRecord, newQuestionRecord.ID).Error; err != nil {
				log.Printf("Failed to add record %v to table %v.\n Reason: %v", newQuestionRecord, &newQuestionRecord, err)
			}
			if newQuestionRecord.ID != 0 {
				log.Printf("Question asked by %v has been stored in the DB", newQuestionRecord.User)
				reply := fmt.Sprintf("Your question has been stored with ID: %v", newQuestionRecord.ID)
				outMsg := qb.rtm.NewOutgoingMessage(reply, newQuestionRecord.SlackChannel)
				qb.rtm.SendMessage(outMsg)
			}
		case newAnswerRecord = <-qb.aCh:
			qb.CreateNewDBRecord(&newAnswerRecord)
			if err := qb.DB.First(&newAnswerRecord, newAnswerRecord.ID).Error; err != nil {
				log.Printf("Failed to add record %v to table %v.\n Reason: %v", newAnswerRecord, &newAnswerRecord, err)
			}
			if newAnswerRecord.ID != 0 {
				log.Printf("An answer has been provided to Question %v by %v", newAnswerRecord.QuestionID, newAnswerRecord.User)
				reply := fmt.Sprintf("Your answer to question %v has been stored with ID: %v", newAnswerRecord.QuestionID, newAnswerRecord.ID)
				outMsg := qb.rtm.NewOutgoingMessage(reply, newAnswerRecord.SlackChannel)
				qb.rtm.SendMessage(outMsg)
			}
		case newUserRecord = <-qb.uCh:
			var count int64
			if err := qb.DB.Where("slack_id = ?", newUserRecord.SlackID).First(&newUserRecord).Count(&count); err != nil { //Count DB entries matching the Slack User ID
				if count == 0 { //Avoid duplicate User entries in the DB.
					qb.CreateNewDBRecord(&newUserRecord)
				}
			}
			if err := qb.DB.First(&newUserRecord, newUserRecord.ID).Error; err != nil {
				log.Printf("Failed to add record %v to table %v.\n Reason: %v", newUserRecord, &newUserRecord, err)
			}
			if newQuestionRecord.ID != 0 {
				log.Printf("Slack User record with %v's information has been stored in the DB", newUserRecord.Name)
			}
		}
	}
}

func (qb *qBot) RunBot() {
	qb.LoadConfig()
	qb.Slack = slack.New(qb.Config.APIToken)
	qb.DB = ConnectToDB()

	//TODO: FIX THIS
	qb.CreateDBTables(models.Question{})
	qb.CreateDBTables(models.Answer{})
	qb.CreateDBTables(models.User{})

	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm

	qb.SetupHandlers()
	qb.rtm.ManageConnection()
}
