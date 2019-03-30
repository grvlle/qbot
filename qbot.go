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
	qListen   chan models.Question
	aListen   chan models.Answer
	msgListen chan Message

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
	qb.qListen = make(chan models.Question, 500)
	qb.aListen = make(chan models.Answer, 500)
	qb.msgListen = make(chan Message, 500)

	return qb
}

func (qb *qBot) SetupHandlers() {
	go qb.EventListener()
	go qb.CommandParser()
	go qb.QuestionHandler()
	go qb.AnswerHandler()
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

			qb.msgListen <- *msg

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

/*CommandParser parses the Slack messages sent by
users for qBot commands and forwards them accordingly
to handlers */
func (qb *qBot) CommandParser() error {

	for msgs := range qb.msgListen {
		message := msgs.Message                    //Message recieved
		sChannel := msgs.Channel                   //Slack Channel where message were sent
		user, err := qb.rtm.GetUserInfo(msgs.User) //User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		msgSplit := []rune(message)
		outMsg := qb.rtm.NewOutgoingMessage("", sChannel)

		switch { //Checks incoming message for requested bot command
		case string(msgSplit[0:3]) == "!q " || string(msgSplit[0:3]) == "!Q ":
			outQuestion := string(msgSplit[3:])
			q := new(models.Question)
			q.User, q.Question, q.SlackChannel = user.Profile.RealName, outQuestion, sChannel
			outMsg = qb.rtm.NewOutgoingMessage("Question stored!", sChannel)

			qb.qListen <- *q

		case string(msgSplit[0:3]) == "!lq" || string(msgSplit[0:3]) == "!LQ":
			outMsg = qb.rtm.NewOutgoingMessage("List Questions", sChannel)
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

			qb.aListen <- *a

		}
		qb.rtm.SendMessage(outMsg)
	}
	return nil
}

/*QuestionHandler stores questions asked in the qBot DB*/
func (qb *qBot) QuestionHandler() {
	for q := range qb.qListen {
		qb.CreateNewDBRecord(&q)
		if err := qb.DB.First(&q, q.ID).Error; err != nil {
			log.Printf("Failed to add record %v to table %v.\n Reason: %v", q, &q, err)
		}
		log.Printf("Question asked by %v has been stored in the DB", q.User)
		reply := fmt.Sprintf("Your question has been stored with ID: %v", q.ID)
		outMsg := qb.rtm.NewOutgoingMessage(reply, q.SlackChannel)
		qb.rtm.SendMessage(outMsg)
	}
}

func (qb *qBot) AnswerHandler() {
	for a := range qb.aListen {
		qb.CreateNewDBRecord(&a)
		if err := qb.DB.First(&a, a.ID).Error; err != nil {
			log.Printf("Failed to add record %v to table %v.\n Reason: %v", a, &a, err)
		}
		log.Printf("An answer has been provided to Question %v by %v", a.QuestionID, a.User)
		reply := fmt.Sprintf("Your answer to question %v has been stored with ID: %v", a.QuestionID, a.ID)
		outMsg := qb.rtm.NewOutgoingMessage(reply, a.SlackChannel)
		qb.rtm.SendMessage(outMsg)

		//Update the questions table with DB ID to the answer as well as answer status
		if err := qb.DB.Table("questions").Where("id IN (?)", a.QuestionID).Updates(map[string]interface{}{"answer_id": a.ID, "answered": true}).Error; err != nil {
			log.Printf("Failed to update the questions table with the answer provided\n Reason: %v", err)
		}

	}
}

func (qb *qBot) RunBot() {
	qb.LoadConfig()
	qb.Slack = slack.New(qb.Config.APIToken)
	qb.DB = ConnectToDB()
	qb.CreateDBTables() //Sets up the DB tables for new Databases

	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm

	qb.SetupHandlers()
	qb.rtm.ManageConnection()
}
