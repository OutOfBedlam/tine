package telegram

import (
	"sync"

	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var bots = map[string]*tgbot.BotAPI{}
var botsLock sync.Mutex

func GetOrNewBot(token string) (*tgbot.BotAPI, error) {
	botsLock.Lock()
	defer botsLock.Unlock()
	if bot, ok := bots[token]; ok {
		return bot, nil
	}
	if bot, err := tgbot.NewBotAPI(token); err != nil {
		return nil, err
	} else {
		bots[token] = bot
		return bot, nil
	}
}
