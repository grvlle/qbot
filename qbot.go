package main

import (
	"fmt"
	"io/ioutil"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
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
	Listener chan *Listener
}

func (qb *qBot) LoadConfig() *qBot {

	qb.Listener = make(chan *Listener, 500)

	content, err := ioutil.ReadFile("config.yaml")
	err = yaml.Unmarshal(content, &qb.Config)
	if err != nil {
		panic(err)
	}
	return qb
}

func (qb *qBot) RunBot() {
	qb.Slack = slack.New(qb.Config.APIToken)

	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm

	qb.SetupHandlers()
	qb.rtm.ManageConnection()
}

func (qb *qBot) SetupHandlers() {
	go qb.EventListener(qb.Listener)
}

func (qb *qBot) EventListener(listen *Listener) {

	for msg := range qb.rtm.IncomingEvents {
		switch ev := msg.Data.(type) {

		case *slack.MessageEvent:
			msg := ev.Text
			fmt.Println(msg)
			listen <- msg
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

func main() {
	var qb qBot
	qb.LoadConfig()
	qb.RunBot()
	//fmt.Print(SlackConfig.APIToken)

}
