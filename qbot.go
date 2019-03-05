package main

import (
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
}

func (qb *qBot) LoadConfig() *qBot {
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
	return
}

func main() {
	var qb qBot
	qb.LoadConfig()
	qb.RunBot()

	//fmt.Print(SlackConfig.APIToken)

}
