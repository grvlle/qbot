package main

import (
	"log"

	"github.com/nlopes/slack"
)

//Reply is used to construct formatted replies
type Reply struct {
	Body        string
	Blocks      []slack.Block
	Attachments []slack.Attachment
	AsUser      bool
}

//ListQuestions takes the Slack Client and returns a formatted
//Reply with unanswered questions to PostFormattedReply
func ListQuestions(client *slack.Client, sChannel string) (string, error) {

	// Format Header Section
	headerText := slack.NewTextBlockObject("mrkdwn", "You have a new request:\n*<fakeLink.toEmployeeProfile.com|Fred Enriquez - New device request>*", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Format Fields
	typeField := slack.NewTextBlockObject("mrkdwn", "*Type:*\nComputer (laptop)", false, false)
	whenField := slack.NewTextBlockObject("mrkdwn", "*When:*\nSubmitted Aut 10", false, false)
	lastUpdateField := slack.NewTextBlockObject("mrkdwn", "*Last Update:*\nMar 10, 2015 (3 years, 5 months)", false, false)
	reasonField := slack.NewTextBlockObject("mrkdwn", "*Reason:*\nAll vowel keys aren't working.", false, false)
	specsField := slack.NewTextBlockObject("mrkdwn", "*Specs:*\n\"Cheetah Pro 15\" - Fast, really fast\"", false, false)

	fieldSlice := make([]*slack.TextBlockObject, 0)
	fieldSlice = append(fieldSlice, typeField)
	fieldSlice = append(fieldSlice, whenField)
	fieldSlice = append(fieldSlice, lastUpdateField)
	fieldSlice = append(fieldSlice, reasonField)
	fieldSlice = append(fieldSlice, specsField)

	fieldsSection := slack.NewSectionBlock(nil, fieldSlice, nil)

	// Include Approve and Deny Buttons
	approveBtnTxt := slack.NewTextBlockObject("plain_text", "Approve", false, false)
	approveBtn := slack.NewButtonBlockElement("", "click_me_123", approveBtnTxt)

	denyBtnTxt := slack.NewTextBlockObject("plain_text", "Deny", false, false)
	denyBtn := slack.NewButtonBlockElement("", "click_me_123", denyBtnTxt)

	actionBlock := slack.NewActionBlock("", approveBtn, denyBtn)

	//Package formatted reply
	var r = &Reply{
		Blocks: []slack.Block{headerSection, fieldsSection, actionBlock},
		AsUser: true,
	}

	return PostFormattedReply(client, sChannel, r)
}

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
