package telegram

import (
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bots = map[string]*tgbot.BotAPI{}

func GetOrNewBot(token string) (*tgbot.BotAPI, error) {
	if bot, ok := bots[token]; ok {
		return bot, nil
	}
	if bot, err := tgbot.NewBotAPI(token); err != nil {
		return nil, err
	} else {
		return bot, nil
	}
}
