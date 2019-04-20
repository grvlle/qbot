package qbot

import (
	"encoding/json"
	"fmt"

	models "github.com/grvlle/qbot/model"
	"github.com/nlopes/slack"
	"github.com/rs/zerolog/log"
)

// Reply is used to construct formatted replies
type Reply struct {
	Body        string
	Blocks      []slack.Block
	Attachments []slack.Attachment
	AsUser      bool
}

// QuestionsAndAnswers are used to parse DB objects into a
// datatype that is easier to work with.
type QuestionsAndAnswers struct {
	QuestionID int    `json:"ID"`
	Question   string `json:"Question"`
	AskedBy    string `json:"UserName"`
	Answers    []struct {
		Answer     string `json:"Answer"`
		AnsweredBy string `json:"UserName"`
	} `json:"Answers"`
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
	return ts, nil
}

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
		reply = "Someone has already asked that question. Run *!lq* to see the last questions asked"
	} else {
		reply = fmt.Sprintf("Thank you %s for providing a question. Your question has been assigned ID: %v", q.UserName, q.ID)
	}
	outMsg := qb.rtm.NewOutgoingMessage(reply, sChannel)
	qb.rtm.SendMessage(outMsg)
}

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
	query, err := qb.DB.QueryQuestions()
	if err != nil {
		log.Error().Err(err)
	}

	ParseQueryAndCacheContent(query, &qStore)
	if len(qStore) > 0 {
		for i := range qStore {
			title := fmt.Sprintf("Question ID %v:", qStore[i].QuestionID)
			footer := fmt.Sprintf("Asked by %s", qStore[i].AskedBy)
			att := []slack.Attachment{slack.Attachment{Color: "#1D9BD1", Title: title, Footer: footer, Text: qStore[i].Question}}
			r.Attachments, r.AsUser = append(att), true
			PostFormattedReply(qb.Slack, sChannel, r)
		}
	}
	qStore = nil // GC
}

// qnaHandler triggers when slack user types !qna <Question ID>.
// Replies to the user with all answers related to the question ID
func (qb *QBot) qnaHandler(sChannel string, questionID int) {
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
				reply := fmt.Sprintf("*Question ID %v* asked by _%v_:\n>%v\n*Answers provided*: \n\n", qnaStore[i].QuestionID, qnaStore[i].AskedBy, qnaStore[i].Question)
				r.Body, r.AsUser = reply, true
				PostFormattedReply(qb.Slack, sChannel, r)

				for _, a := range qnaStore[i].Answers {
					att := []slack.Attachment{slack.Attachment{Color: "#36a64f", Footer: a.AnsweredBy, Text: a.Answer}}
					r2.Attachments, r2.AsUser = append(att), true
					PostFormattedReply(qb.Slack, sChannel, r2)

				}
			}

		}

	}
	qnaStore = nil // GC
}

// ParseQueryAndCacheContent takes query data (db object) and a &struct (buffer) and populates
// it using json tags.
func ParseQueryAndCacheContent(data, buffer interface{}) error {
	jsonEncQNA, _ := json.Marshal(data)
	return json.Unmarshal(jsonEncQNA, &buffer)
}
