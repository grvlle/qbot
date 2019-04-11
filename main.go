package main

import (
	qb "github.com/grvlle/qbot/slackbot"
)

var bot qb.Bot

func init() {
	InitializeLogger()
}

func main() {
	bot.RunBot()
}
