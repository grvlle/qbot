package qbot

import (
	"log"

	"github.com/nlopes/slack"
)

// Reply is used to construct formatted replies
type Reply struct {
	Body        string
	Blocks      []slack.Block
	Attachments []slack.Attachment
	AsUser      bool
}

type QuestionsAndAnswers struct {
	QuestionID int    `json:"ID"`
	Question   string `json:"Question"`
	AskedBy    string `json:"UserName"`
	Answers    []struct {
		Answer     string `json:"Answer"`
		AnsweredBy string `json:"UserName"`
	} `json:"Answers"`
}

// ListQuestions takes the Slack Client and returns a formatted
// Reply with unanswered questions to PostFormattedReply
// func (qb *QBot) ListQuestions(client *slack.Client, sChannel string) (string, error) {

//ltq := qb.DB.LastTenQuestions(&db.LastTenQuestions{})
//lta, _ := qb.LastTenAnswers(&LastTenAnswers{})

// for i, q := range ltq.ID {
// 	if int(q) == lta.QuestionID[i] {
// 		fmt.Println("Matching")
// 		fmt.Println(ltq.Question[q])
// 	}
// }

// TODO: Come up with a better solution not to crash the program at less than 3 questions asked
//if len(ltq.ID) >= 3 {

// // Shared Assets
// divSection := slack.NewDividerBlock()
// voteBtnText := slack.NewTextBlockObject("plain_text", "Answer", true, false)
// voteBtnEle := slack.NewButtonBlockElement("", "click_me_123", voteBtnText)
// intBtnText := slack.NewTextBlockObject("plain_text", "Curious?", true, false)
// intBtnEle := slack.NewButtonBlockElement("", "click_me_123", intBtnText)
// profileOne := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_1.png", "Michael Scott")
// profileTwo := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_2.png", "Dwight Schrute")

// // Header Section
// headerText := slack.NewTextBlockObject("mrkdwn", ":information_source: *Listing the three most recent questions asked*", false, false)
// headerSection := slack.NewSectionBlock(headerText, nil, nil)

// // Option One Info
// optOneText := slack.NewTextBlockObject("mrkdwn", ":small_blue_diamond: *Question "+strconv.Itoa(int(ltq.ID[0]))+":* "+ltq.Question[0], false, false)
// optOne2Text := slack.NewTextBlockObject("mrkdwn", "Asked by *<fakeLink.toYourApp.com|Martin Granstrom>*", false, false)
// optOne3Text := slack.NewTextBlockObject("plain_text", "2 netrounders are curious", false, false)
// optOneContext := slack.NewContextBlock("", profileOne, profileTwo, optOne3Text)
// optOneSection := slack.NewSectionBlock(optOneText, nil, voteBtnEle)
// optOne1Section := slack.NewSectionBlock(optOne2Text, nil, intBtnEle)

// // Option Two Info
// optTwoText := slack.NewTextBlockObject("mrkdwn", ":small_blue_diamond: *Question "+strconv.Itoa(int(ltq.ID[1]))+":* "+ltq.Question[1], false, false)
// optTwo2Text := slack.NewTextBlockObject("mrkdwn", "Asked by *<fakeLink.toYourApp.com|Miguel Gomez>*", false, false)
// optTwo3Text := slack.NewTextBlockObject("plain_text", "1 netrounders are curious", false, false)
// optTwoContext := slack.NewContextBlock("", profileTwo, optTwo3Text)
// optTwoSection := slack.NewSectionBlock(optTwoText, nil, voteBtnEle)
// optTwo1Section := slack.NewSectionBlock(optTwo2Text, nil, intBtnEle)

// // Option Three Info
// optThreeText := slack.NewTextBlockObject("mrkdwn", ":small_blue_diamond: *Question "+strconv.Itoa(int(ltq.ID[2]))+":* "+ltq.Question[2], false, false)
// optThree2Text := slack.NewTextBlockObject("mrkdwn", "Asked by *<fakeLink.toYourApp.com|John Doe>*", false, false)
// optThree3Text := slack.NewTextBlockObject("plain_text", "0 netrounders are curious", false, false)
// optThreeContext := slack.NewContextBlock("", optThree3Text)
// optThreeSection := slack.NewSectionBlock(optThreeText, nil, voteBtnEle)
// optThree1Section := slack.NewSectionBlock(optThree2Text, nil, intBtnEle)

// // Suggestions Action
// btnTxt := slack.NewTextBlockObject("plain_text", "Next 3 Questions >", false, false)
// nextBtn := slack.NewButtonBlockElement("", "click_me_123", btnTxt)
// actionBlock := slack.NewActionBlock("", nextBtn)

// //Package formatted reply
// var r = &Reply{
// 	Blocks: []slack.Block{
// 		headerSection,
// 		divSection,
// 		optOneSection,
// 		optOne1Section,
// 		divSection,
// 		optOneContext,
// 		divSection,
// 		optTwoSection,
// 		optTwo1Section,
// 		divSection,
// 		optTwoContext,
// 		divSection,
// 		optThreeSection,
// 		optThree1Section,
// 		divSection,
// 		optThreeContext,
// 		divSection,
// 		actionBlock},
// 	AsUser: true,
// }

// return PostFormattedReply(client, sChannel, r)
// }
// return "", nil //TODO: Fix this function to not crash at > 3 questions
// }

//PostFormattedReply takes a pointer to the Slack Client and a
//Reply and posts it to the requesting channel
func PostFormattedReply(client *slack.Client, sChannel string, r *Reply) (string, error) {
	_, ts, err := client.PostMessage(
		sChannel,
		slack.MsgOptionText(r.Body, false),
		slack.MsgOptionBlocks(r.Blocks...),
		slack.MsgOptionAttachments(r.Attachments...),
		slack.MsgOptionAsUser(r.AsUser),
		slack.MsgOptionEnableLinkUnfurl(),
	)
	if err != nil {
		log.Printf("Unable to Post Message to channel\n Reason: %v", err)
	}
	return ts, nil
}
