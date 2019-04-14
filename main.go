package main

import (
	qb "github.com/grvlle/qbot/qbot"
)

var bot qb.QBot

func init() {
	InitializeLogger()
}

func main() {
	bot.RunBot()
}
