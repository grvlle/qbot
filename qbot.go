package main

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

var (
	bot       qBot
	messageCh = make(chan Message, 500)
)

type qBot struct {
	// Global qBot configuration
	Config struct {
		APIToken       string   `yaml:"apiToken"`
		JoinChannels   []string `yaml:"joinChannels"`
		GeneralChannel string   `yaml:"generalChannel"`
		Debug          bool
	}

	//Establish connection
	Slack *slack.Client
	rtm   *slack.RTM
	Users map[string]slack.User
	//Channels map[string]Channel

	//Listeners
	WG sync.WaitGroup
}

func (qb *qBot) LoadConfig() *qBot {

	content, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(content, &qb.Config)
	if err != nil {
		panic(err)
	}
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
	go qb.EventReciever()
}

func (qb *qBot) EventListener() {
	defer qb.WG.Done()

	for events := range qb.rtm.IncomingEvents {
		switch ev := events.Data.(type) {

		case *slack.MessageEvent:
			msg := new(Message)
			msg.User, msg.Channel, msg.Message = ev.Username, ev.Channel, ev.Text
			messageCh <- *msg

		case *slack.ConnectedEvent:
			fmt.Println("Infos:", ev.Info)
			fmt.Println("Connection counter:", ev.ConnectionCount)
			// Replace C2147483705 with your Channel ID
			//qb.rtm.SendMessage(qb.rtm.NewOutgoingMessage("Hello world", "C2147483705"))

		case *slack.PresenceChangeEvent:
			fmt.Printf("Presence Change: %v\n", ev)

		case *slack.LatencyReport:
			fmt.Printf("Current latency: %v\n", ev.Value)

		case *slack.RTMError:
			fmt.Printf("Error: %s\n", ev.Error())

		case *slack.InvalidAuthEvent:
			fmt.Printf("Invalid credentials")
			return
		}
	}
}

func (qb *qBot) EventReciever() {
	for msgs := range messageCh {
		fmt.Println(msgs.Message)
	}
	// content := map[string]string{
	// 	"username": username,
	// 	"channel":  channel,
	// 	"message":  msg,
	// }
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
	bot.RunBot()
}
