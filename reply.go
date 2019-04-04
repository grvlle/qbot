package main

import (
	"fmt"
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
func (qb *qBot) ListQuestions(client *slack.Client, sChannel string) (string, error) {

	fmt.Println(qb.LastTenQuestions(&LastTenQuestions{}))
	fmt.Println(qb.LastTenAnswers(&LastTenAnswers{}))

	// Shared Assets for example
	divSection := slack.NewDividerBlock()
	voteBtnText := slack.NewTextBlockObject("plain_text", "Vote", true, false)
	voteBtnEle := slack.NewButtonBlockElement("", "click_me_123", voteBtnText)
	profileOne := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_1.png", "Michael Scott")
	profileTwo := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_2.png", "Dwight Schrute")
	profileThree := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_3.png", "Pam Beasely")
	profileFour := slack.NewImageBlockObject("https://api.slack.com/img/blocks/bkb_template_images/profile_4.png", "Angela")

	// Header Section
	headerText := slack.NewTextBlockObject("mrkdwn", "*Where should we order lunch from?* Poll by <fakeLink.toUser.com|Mark>", false, false)
	headerSection := slack.NewSectionBlock(headerText, nil, nil)

	// Option One Info
	optOneText := slack.NewTextBlockObject("mrkdwn", ":sushi: *Ace Wasabi Rock-n-Roll Sushi Bar*\nThe best landlocked sushi restaurant.", false, false)
	optOneSection := slack.NewSectionBlock(optOneText, nil, voteBtnEle)

	// Option One Votes
	optOneVoteText := slack.NewTextBlockObject("plain_text", "3 votes", true, false)
	optOneContext := slack.NewContextBlock("", profileOne, profileTwo, profileThree, optOneVoteText)

	// Option Two Info
	optTwoText := slack.NewTextBlockObject("mrkdwn", ":hamburger: *Super Hungryman Hamburgers*\nOnly for the hungriest of the hungry.", false, false)
	optTwoSection := slack.NewSectionBlock(optTwoText, nil, voteBtnEle)

	// Option Two Votes
	optTwoVoteText := slack.NewTextBlockObject("plain_text", "2 votes", true, false)
	optTwoContext := slack.NewContextBlock("", profileFour, profileTwo, optTwoVoteText)

	// Option Three Info
	optThreeText := slack.NewTextBlockObject("mrkdwn", ":ramen: *Kagawa-Ya Udon Noodle Shop*\nDo you like to shop for noodles? We have noodles.", false, false)
	optThreeSection := slack.NewSectionBlock(optThreeText, nil, voteBtnEle)

	// Option Three Votes
	optThreeVoteText := slack.NewTextBlockObject("plain_text", "No votes", true, false)
	optThreeContext := slack.NewContextBlock("", optThreeVoteText)

	// Suggestions Action
	btnTxt := slack.NewTextBlockObject("plain_text", "Add a suggestion", false, false)
	nextBtn := slack.NewButtonBlockElement("", "click_me_123", btnTxt)
	actionBlock := slack.NewActionBlock("", nextBtn)

	//Package formatted reply
	var r = &Reply{
		Blocks: []slack.Block{
			headerSection,
			divSection,
			optOneSection,
			optOneContext,
			optTwoSection,
			optTwoContext,
			optThreeSection,
			optThreeContext,
			divSection,
			actionBlock},
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
