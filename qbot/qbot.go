package qbot

import (
	"fmt"
	"io/ioutil"

	"strconv"
	"strings"

	db "github.com/grvlle/qbot/db"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// RunBot will initiate the bot
func (qb *QBot) RunBot() {
	qb.LoadConfig()
	qb.Slack = slack.New(qb.Config.APIToken)
	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm
	qb.SetupHandlers()
	qb.rtm.ManageConnection()

}

// QBot contains Slack API configuration data
// And provides Websocket and DB access
type QBot struct {
	// Global QBot configuration
	Config struct {
		APIToken string `yaml:"apiToken"`
		Metadata struct {
			JoinChannels   string `yaml:"joinChannels"`
			GeneralChannel string `yaml:"generalChannel"`
		}
		DebugLevel string `yaml:"debug"`
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
// will populate the Config struct in the QBot type
// with configuration variables
func (qb *QBot) LoadConfig() *QBot {
	content, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(content, &qb.Config)
	if err != nil {
		panic(err)
	}
	qb.msgCh = make(chan Message, 500)

	// qb.DB = db.InitializeDB()

	return qb
}

// SetupHandlers sets up the Go Routines
func (qb *QBot) SetupHandlers() {
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
func (qb *QBot) EventListener() {
	for events := range qb.rtm.IncomingEvents {
		switch ev := events.Data.(type) {
		case *slack.MessageEvent:
			r := new(Message)
			r.User, r.Channel, r.Message = ev.User, ev.Channel, ev.Text
			qb.msgCh <- *r
		case *slack.ConnectedEvent:
			log.Info().Msgf("Info: %v", *ev.Info)
			log.Info().Msgf("Connection counter: %v", ev.ConnectionCount)
		case *slack.PresenceChangeEvent:
			log.Info().Msgf("Presence Change: %v\n", ev)
		case *slack.LatencyReport:
			//fmt.Printf("Current latency: %v\n", ev.Value)
		case *slack.RTMError:
			log.Error().Msgf("RTM Error: %s\n", ev.Error())
		case *slack.InvalidAuthEvent:
			log.Warn().Msg("Invalid credentials")
			return
		}
	}
}

// CommandParser parses the Slack messages for QBot commands
func (qb *QBot) CommandParser() {
	for msgs := range qb.msgCh {
		message := msgs.Message                        // Message recieved
		sChannel := msgs.Channel                       // Slack Channel where message were sent
		userInfo, err := qb.rtm.GetUserInfo(msgs.User) // User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}

		msgSplit := []rune(message)

		switch { // Checks incoming message for requested bot command

		// Ask Questions
		case string(msgSplit[0:3]) == "!q " || string(msgSplit[0:3]) == "!Q ":
			question := string(msgSplit[3:])
			go qb.qHandler(sChannel, question, userInfo)

		// Answer Questions
		case string(msgSplit[0:3]) == "!a " || string(msgSplit[0:3]) == "!A ":
			// TODO: Capture error wher users doesn't include an ID
			var reply string
			msg := string(msgSplit[3:])
			parts := strings.Fields(msg)     // Splits incoming message into slice
			questionID, err := idParser(msg) // Verifies that the first element after "!a " is an intiger (Question ID)
			if err != nil {
				log.Warn().Msg("Question ID was not provided with the question answered")
				reply = fmt.Sprintf("Please include an ID for the question you're answering\n E.g '!a 123 The answer is no!'")
			}
			answer := strings.Join(parts[1:], " ")
			if len(answer) != 0 {
				go qb.aHandler(sChannel, answer, questionID, userInfo)
			} else {
				reply = "No answer was provided, please try again"
				log.Warn().Msg("Slack user failed to provide and answer")
			}
			qb.rtm.SendMessage(qb.rtm.NewOutgoingMessage(reply, sChannel))

		// List Questions
		case string(msgSplit[0:3]) == "!lq" || string(msgSplit[0:3]) == "!LQ":
			go qb.lqHandler(sChannel)

		// List Answers
		case string(msgSplit[0:4]) == "!la " || string(msgSplit[0:4]) == "!LA ":
			// TODO: Capture error wher users doesn't include an ID
			var reply string
			msg := string(msgSplit[4:])
			questionID, err := idParser(msg)
			if err != nil {
				log.Warn().Msg("Question ID was not provided when listing answers")
				reply = fmt.Sprintf("Please include an ID for the question you're trying to list the answers for\n E.g '!la 123'")
			}
			qb.rtm.SendMessage(qb.rtm.NewOutgoingMessage(reply, sChannel))
			go qb.laHandler(sChannel, questionID)

		// Delete Question
		case string(msgSplit[0:10]) == "!delete_q ":
			var reply string
			msg := string(msgSplit[10:])
			questionID, _ := idParser(msg)
			if err != nil {
				log.Warn().Msg("Question ID was not provided when trying to delete question")
				reply = fmt.Sprintf("Please include an ID for the question you're trying to delete\n E.g '!delete_q 123'")
			}
			qb.rtm.SendMessage(qb.rtm.NewOutgoingMessage(reply, sChannel))
			go qb.deleteqHandler(sChannel, userInfo.ID, questionID)

		// Help Information
		case string(msgSplit[0:2]) == "!h" || string(msgSplit[0:5]) == "!help":
			qb.helpHandler(sChannel)
		}
	}
}

// Verifies that the first element is an intiger and returns it
func idParser(message string) (int, error) {
	questionID, err := strconv.Atoi(strings.Fields(message)[0])
	return questionID, err
}
