package qbot

import (
	"fmt"
	"io/ioutil"

	"github.com/nlopes/slack"
	"gopkg.in/yaml.v2"
)

type SlackConfig struct {
	APIToken       string   `yaml:"apiToken"`
	JoinChannels   []string `yaml:"joinChannels"`
	GeneralChannel string   `yaml:"generalChannel"`
	Debug          bool
}

type qBot struct {
	// Global qBot configuration
	Config SlackConfig

	//Establish connection
	Slack *slack.Client
	rtm   *slack.RTM
	Users map[string]slack.User
	//Channels map[string]Channel
}

func loadConfig() {
	var sc SlackConfig

	content, err := ioutil.ReadFile("config.yaml")

	err = yaml.Unmarshal(content, &sc)
	if err != nil {
		panic(err)
	}
	fmt.Print(sc.APIToken)
}

func (qb *qBot) runBot() {
	qb.Slack = slack.New(qb.Config.APIToken)

	rtm := qb.Slack.NewRTM()
	qb.rtm = rtm

	//qb.setupHandlers()

	qb.rtm.ManageConnection()
}

func main() {
	loadConfig()
	//fmt.Print(SlackConfig.APIToken)

}
