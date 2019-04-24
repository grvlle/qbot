package qbot

import (
	"encoding/json"
	"fmt"
	"time"

	models "github.com/grvlle/qbot/model"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
)

// Reply is used to construct formatted replies
type Reply struct {
	Body        string
	Blocks      []slack.Block      // https://api.slack.com/tools/block-kit-builder
	Attachments []slack.Attachment // https://api.slack.com/docs/message-formatting
	AsUser      bool
}

// UserInformation are used to parse DB objects into a
// datatype that is easier to work with. The ParseQueryAndCacheContent
// function is used to populate the below struct.
type UserInformation struct {
	UserID    int `json:"id,omitempty"`
	Questions []struct {
		QuestionID int    `json:"id,omitempty"`
		Question   string `json:"question,omitempty"`
	} `json:"questions,omitempty"`
	Answers []struct {
		AnswerID int    `json:"id,omitempty"`
		Answer   string `json:"answer,omitempty"`
	} `json:"answers,omitempty"`
}

// QuestionsAndAnswers are used to parse DB objects into a
// datatype that is easier to work with. The ParseQueryAndCacheContent
// function is used to populate the below struct.
type QuestionsAndAnswers struct {
	QuestionID int    `json:"id,omitempty"`
	Question   string `json:"question,omitempty"`
	AskedBy    string `json:"username,omitempty"`
	CreatedAt  string `json:"createdat,omitempty"`
	Answers    []struct {
		Answer     string `json:"answer,omitempty"`
		AnsweredBy string `json:"username,omitempty"`
		CreatedAt  string `json:"createdat,omitempty"`
	} `json:"answers,omitempty"`
}

// PostFormattedReply takes a pointer to the Slack Client and a
// Reply and posts it to the requesting channel
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
		log.Error().Msgf("Unable to Post Message to channel: %v", err.Error())
	}
	return ts, err
}

// qHandler triggers when a slack user uses !q to provide a question.
// It will update the database with the question and add non-existing users as well.
func (qb *QBot) qHandler(sChannel, outQuestion string, userInfo *slack.User) {
	var reply string
	q := new(models.Question)
	q.Question, q.SlackChannel, q.UserName = outQuestion, sChannel, userInfo.Profile.RealName
	user := &models.User{
		Questions: []*models.Question{q},
		Name:      userInfo.Profile.RealNameNormalized,
		Title:     userInfo.Profile.Title,
		Avatar:    userInfo.Profile.Image32,
		SlackUser: userInfo.ID,
	}
	if !qb.DB.UserExistInDB(*user) {
		qb.DB.UpdateUsers(user)
	}
	// Update the user_questions (m2m) and questions table with question_id and question
	if err := qb.DB.UpdateUserTableWithQuestion(user, q); err != nil {
		log.Error().Err(err)
		reply = "Either the question exceeds 300 words, or someone has already asked that question. Run *!lq* to see the last questions asked"
	} else {
		reply = fmt.Sprintf("Thank you %s for providing a question. Your question has been assigned ID: %v", q.UserName, q.ID)
	}
	outMsg := qb.rtm.NewOutgoingMessage(reply, sChannel)
	qb.rtm.SendMessage(outMsg)
}

// aHandler triggers when a slack user uses !a to provide an answer.
// It will update the database with the answer and add non-existing users as well.
func (qb *QBot) aHandler(sChannel, outAnswer string, questionID int, userInfo *slack.User) {
	var reply string
	a := new(models.Answer)
	a.Answer, a.QuestionID, a.SlackChannel, a.UserName = outAnswer, questionID, sChannel, userInfo.Profile.RealName
	user := &models.User{
		Answers:   []*models.Answer{a},
		Name:      userInfo.Profile.RealNameNormalized,
		Title:     userInfo.Profile.Title,
		Avatar:    userInfo.Profile.Image32,
		SlackUser: userInfo.ID,
	}
	if !qb.DB.UserExistInDB(*user) {
		qb.DB.UpdateUsers(user)
	}
	// Update the user_answers (m2m) and answers table with the answer_id and answer
	if err := qb.DB.UpdateUserTableWithAnswer(user, a); err != nil {
		log.Error().Err(err)
		reply = "I had problems storing your provided answer in the DB"
	}
	// Update the questions_answer (m2m) table record with the answer_id
	q := models.Question{Answers: []*models.Answer{a}}
	if err := qb.DB.UpdateQuestionTableWithAnswer(&q, a); err != nil {
		log.Error().Err(err)
		reply = "I had problems storing your provided answer in the DB. Did you specify the Question ID correctly?"
	} else {
		reply = fmt.Sprintf("Thank you %s for providing an answer to question %v!", a.UserName, q.ID)
	}
	outMsg := qb.rtm.NewOutgoingMessage(reply, sChannel)
	qb.rtm.SendMessage(outMsg)
}

// lqHandler triggers when a slack user types !lq. lqHandler proceeds
// to construct a formatted reply listing a restricted query set of all
// questions asked.
func (qb *QBot) lqHandler(sChannel string) {
	r := new(Reply)
	var qStore []QuestionsAndAnswers
	query, err := qb.DB.QueryAnsweredQuestions()
	if err != nil {
		log.Error().Err(err)
	}
	ParseQueryAndCacheContent(query, &qStore)
	PostFormattedReply(qb.Slack, sChannel, &Reply{Body: "Below is a list of the five most recent questions asked. The green color marks answered questions. Use `!la <Question ID>` to list the answers.", AsUser: true})
	if len(qStore) > 0 {
		for i := range qStore {
			ts, _ := time.Parse(time.RFC3339, qStore[i].CreatedAt)
			title := fmt.Sprintf("Question ID %v:", qStore[i].QuestionID)
			footer := fmt.Sprintf("Asked by %s | %s", qStore[i].AskedBy, ts.String())
			att := []slack.Attachment{slack.Attachment{Color: "#1D9BD1", Title: title, Footer: footer, Text: qStore[i].Question}}
			if len(qStore[i].Answers) >= 1 { // If question is answered, the output will be colored green
				att = []slack.Attachment{slack.Attachment{Color: "#36a64f", Title: title, Footer: footer, Text: qStore[i].Question}}
			}
			r.Attachments, r.AsUser = append(att), true
			PostFormattedReply(qb.Slack, sChannel, r)
		}
	}
}

// laHandler triggers when slack user types !la <Question ID>.
// Replies to the user with all answers related to the question ID
func (qb *QBot) laHandler(sChannel string, questionID int) {
	r := new(Reply)
	r2 := new(Reply)
	var qnaStore []QuestionsAndAnswers
	query, err := qb.DB.QueryAnsweredQuestionsByID(questionID)
	if err != nil {
		log.Error().Err(err)
	}

	ParseQueryAndCacheContent(query, &qnaStore)
	if len(qnaStore) > 0 {
		for i := range qnaStore {
			if len(qnaStore[i].Answers) >= 1 {
				reply := fmt.Sprintf("*Question ID %v* asked by _%v_:\n\n", qnaStore[i].QuestionID, qnaStore[i].AskedBy)
				att := []slack.Attachment{slack.Attachment{Color: "#1D9BD1", Pretext: reply, Text: qnaStore[i].Question}, slack.Attachment{Pretext: "\n*Answers Provided:*"}}
				r.Attachments, r.AsUser = att, true
				PostFormattedReply(qb.Slack, sChannel, r)

				for _, a := range qnaStore[i].Answers {
					ts, _ := time.Parse(time.RFC3339, a.CreatedAt)
					footer := fmt.Sprintf("Answered by %s | %s", a.AnsweredBy, ts)
					att := []slack.Attachment{slack.Attachment{Color: "#36a64f", Footer: footer, Text: a.Answer}}
					r2.Attachments, r2.AsUser = append(att), true
					PostFormattedReply(qb.Slack, sChannel, r2)
				}
			} else {
				PostFormattedReply(qb.Slack, sChannel, &Reply{Body: "This question has not been answered yet. To provide an answer use `!a <Question ID> <Answer>`", AsUser: true})
			}
		}
	}
}

// deleteqHandler will soft delete user posted questions and return a reply.
func (qb *QBot) deleteqHandler(sChannel, sUserID string, questionID int) (string, error) {
	var reply string
	var userStore []UserInformation
	query, err := qb.DB.QueryQuestionsAskedByUserID(sUserID)
	if err != nil {
		log.Error().Err(err)
	}
	ParseQueryAndCacheContent(query, &userStore)
	for i := range userStore {
		for _, q := range userStore[i].Questions {
			switch {
			case questionID == q.QuestionID: // Checks if user owns the question
				if err := qb.DB.DeleteAnsweredQuestionsByID(questionID); err != nil {
					log.Error().Err(err)
				}
				reply = "Question has successfully been deleted."
				return PostFormattedReply(qb.Slack, sChannel, &Reply{Body: reply, AsUser: true})
			}
		}
	}
	reply = "Either the question doesn't exist, or you're trying to delete someone elses question. The latter is not allowed. :eyes:"
	return PostFormattedReply(qb.Slack, sChannel, &Reply{Body: reply, AsUser: true})

}

func (qb *QBot) helpHandler(sChannel string) (string, error) {
	r := new(Reply)
	title := ":information_source: HOW TO USE QBOT?"
	text := fmt.Sprintf("Thanks for using the Netrounds qBot! Below is a list of all available bot commands.\n\n" +
		"· *Help:* `!help` will display the Help Information you're looking at right now.\n" +
		"· *Ask Question:* `!q <question>` is used when asking a question.\n" +
		"· *List Questions:* `!lq` will list the last 5 questions asked. Each question will have a color marking indicating wheter or not they have been answered. Blue (:blue_heart:) is an unanswered question, and green (:green_heart:) indicates that atleast one answer has been provided. \n" +
		"· *Answer Questions:* `!a <question ID> <answer>` is used when providing and answer to a question.\n" +
		"· *List Answers:* `!la <question ID>` will list the answers provided to a specific question.\n" +
		"· *Delete Question:* `!delete_q <question ID>` - will delete questions asked along with associated answers.\n\n\n" +
		"More commands and further functionality will be introduced over time. For feature requests and bug reports, please DM <@martin.g>.\n")
	footer := "qBot v.1.0 BETA"
	fields := []slack.AttachmentField{slack.AttachmentField{Title: "Website", Value: "Coming soon...", Short: true}, slack.AttachmentField{Title: "Contribute", Short: true, Value: "This is a hobby project written in Go by <@martin.g>. Feel free to <https://github.com/grvlle/qbot/tree/develop|contribute>! :golang:"}}
	att := []slack.Attachment{slack.Attachment{Color: "#1D9BD1", Title: title, Text: text, Fields: fields, Footer: footer}}
	r.Attachments, r.AsUser = append(att), true
	return PostFormattedReply(qb.Slack, sChannel, r)
}

// ParseQueryAndCacheContent takes query data (db object) and a &struct (buffer) and populates
// it using json tags.
func ParseQueryAndCacheContent(data, buffer interface{}) error {
	jsonEncoded, _ := json.Marshal(data)
	return json.Unmarshal(jsonEncoded, &buffer)
}
