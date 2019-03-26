package main

import (
	"fmt"
	"io/ioutil"
	"time"

	db "./db"
	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

var bot qBot

type qBot struct {
	// Global qBot configuration
	Config struct {
		APIToken       string   `yaml:"apiToken"`
		JoinChannels   []string `yaml:"joinChannels"`
		GeneralChannel string   `yaml:"generalChannel"`
		Debug          bool
	}

	//Establish websocket
	Slack *slack.Client
	rtm   *slack.RTM

	//IO flow
	qListen   chan Question
	msgListen chan Message
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

	qb.qListen = make(chan Question, 500)
	qb.msgListen = make(chan Message, 500)

	return qb
}

/*Message contains the details of a recieved
Slack message. Constructed in the EventListener
method and passed in the messageCh*/
type Message struct {
	User    string
	Channel string
	Message string
}

func (qb *qBot) SetupHandlers() {
	go qb.EventListener()
	go qb.AskQuestion()
	go qb.CommandParser()
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
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			//qb.rtm.SendMessage(qb.rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			//fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return
		}
	}
}

/*Question type will be used to store questions
asked by users in the Database */
type Question struct {
	User         string
	Question     string
	Answered     bool
	ID           int
	SlackChannel string
	TimeStamp    string
}

/*AskedQuestions TODO*/
type AskedQuestions struct {
	Questions []string
}

func (qb *qBot) CommandParser() {

	for msgs := range qb.msgListen {
		message := msgs.Message                    //Message recieved
		sChannel := msgs.Channel                   //Slack Channel where message were sent
		user, err := qb.rtm.GetUserInfo(msgs.User) //User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		msgSplit := []rune(message)
		outMsg := qb.rtm.NewOutgoingMessage("nil", sChannel)

		switch { //Checks incoming message for requested bot command
		case string(msgSplit[0:3]) == "!q " || string(msgSplit[0:3]) == "!Q ":
			outQuestion := string(msgSplit[3:])
			q := new(Question)
			ts := time.Now()
			q.User, q.TimeStamp, q.Question, q.Answered, q.ID, q.SlackChannel = user.Profile.RealName, ts.Format("20060102150405"), outQuestion, false, 1, sChannel
			outMsg = qb.rtm.NewOutgoingMessage("Question stored!", sChannel)

			qb.qListen <- *q

		case string(msgSplit[0:3]) == "!lq" || string(msgSplit[0:3]) == "!LQ":
			outMsg = qb.rtm.NewOutgoingMessage("List questions", sChannel)
		case string(msgSplit[0:4]) == "!qna" || string(msgSplit[0:4]) == "!QnA":
			outMsg = qb.rtm.NewOutgoingMessage("List answer and questions", sChannel)
		case string(msgSplit[0:3]) == "!a " || string(msgSplit[0:3]) == "!A ":
			outMsg = qb.rtm.NewOutgoingMessage("Answer Question", sChannel)
		}
		//fmt.Printf("Channel: %s\n User: %s\n msg: %s\n", sChannel, user.Profile.RealName, message)
		qb.rtm.SendMessage(outMsg)
	}
}

// func AnswerQuestion() {
// 	return "answer"
// }

// func ListQnA() {
// 	return "qna"
// }

//AskQuestion TODO: Store questions asked in DB
func (qb *qBot) AskQuestion() {
	id := 0
	for q := range qb.qListen {
		id++
		q.ID = id //assign incremented ID to incoming question
		fmt.Println(q)
	}
}

func (qb *qBot) RunBot() {
	qb.LoadConfig()
	qb.Slack = slack.New(qb.Config.APIToken)

	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm

	qb.SetupHandlers()
	qb.rtm.ManageConnection()
}

func main() {
	//bot.RunBot()
	db.ConnectToDB()
}
