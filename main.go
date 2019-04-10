package main

import (
	qb "github.com/grvlle/qbot/slackbot"
)

var bot qb.Bot

func main() {
	//qb.RunBot()
	bot.RunBot()
}

// docker exec -it mysql1 mysql -uroot -p
