package main

import (
	"fmt"
	"io/ioutil"

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

	go qb.MessageReciever()
}

/* EventListener listens on the websocket for
incoming slack events, including messages that it
passes to the messageCh channel monitored by
MessageReciever() */
func (qb *qBot) EventListener(messageCh chan<- Message) {

	for events := range qb.rtm.IncomingEvents {
		switch ev := events.Data.(type) {

		case *slack.MessageEvent:
			msg := new(Message)
			msg.User, msg.Channel, msg.Message = ev.User, ev.Channel, ev.Text
			messageCh <- *msg

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

func (qb *qBot) MessageReciever() {
	messageCh := make(chan Message, 500)
	go qb.EventListener(messageCh)

	for msgs := range messageCh {
		message := msgs.Message                    //Message recieved
		schannel := msgs.Channel                   //Channel where message were sent
		user, err := qb.rtm.GetUserInfo(msgs.User) //User that sent message
		if err != nil {
			fmt.Printf("%s\n", err)
		}
		question := []rune(message)
		outmsg := qb.rtm.NewOutgoingMessage("nil", schannel)

		switch {
		case string(question[0:3]) == "!q ":
			outmsg = qb.rtm.NewOutgoingMessage("List questions", schannel)
		case string(question[0:4]) == "!qna":
			outmsg = qb.rtm.NewOutgoingMessage("List answer and questions", schannel)
		case string(question[0:3]) == "!a ":
			outmsg = qb.rtm.NewOutgoingMessage("Answer Question", schannel)
		}
		//fmt.Printf("Channel: %s\n User: %s\n msg: %s\n", schannel, user.Profile.RealName, message)
		fmt.Println(user)
		qb.rtm.SendMessage(outmsg)
	}
}

// func AnswerQuestion() {
// 	return "answer"
// }

// func ListQnA() {
// 	return "qna"
// }

//AskQuestion asfas
// func (qb *qBot) AskQuestion() {
// 	for question, schannel := range qb.qCh {
// 		fmt.Println(question, schannel)
// 	}
// 	outmsg := qb.rtm.NewOutgoingMessage(question, channel)
// 	qb.rtm.SendMessage(outmsg)
// }

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
