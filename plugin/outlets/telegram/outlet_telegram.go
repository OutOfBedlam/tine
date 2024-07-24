package telegram

import (
	tg "github.com/OutOfBedlam/tine/drivers/telegram"
	"github.com/OutOfBedlam/tine/engine"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	engine.RegisterOutlet(&engine.OutletReg{
		Name:    "telegram",
		Factory: TelegramOutlet,
	})
}

func TelegramOutlet(ctx *engine.Context) engine.Outlet {
	return &telegramOutlet{ctx: ctx}
}

type telegramOutlet struct {
	ctx    *engine.Context
	bot    *tgbot.BotAPI
	chatId int64
}

var _ = engine.Outlet((*telegramOutlet)(nil))

func (to *telegramOutlet) Open() error {
	token := to.ctx.Config().GetString("token", "")

	if bot, err := tg.GetOrNewBot(token); err != nil {
		return err
	} else {
		to.bot = bot
		to.bot.Debug = to.ctx.Config().GetBool("debug", false)
		to.chatId = to.ctx.Config().GetInt64("chat_id", 0)
	}
	return nil
}

func (to *telegramOutlet) Close() error {
	return nil
}

func (to *telegramOutlet) Handle(record []engine.Record) error {
	for _, r := range record {
		var chatId int64
		idField := r.Field("chat_id")
		if idField != nil && idField.Type() == engine.INT && !idField.IsNull() {
			chatId, _ = idField.Value.Int64()
		} else {
			if to.chatId == 0 {
				to.ctx.LogDebug("outlets.telegram", "invalid record no chat_id", r)
				continue
			} else {
				chatId = to.chatId
			}
		}
		textField := r.Field("text")
		if textField == nil || textField.Type() != engine.STRING || textField.IsNull() {
			to.ctx.LogDebug("outlets.telegram", "invalid record no text", r)
			continue
		}
		text, _ := textField.Value.String()
		if text == "" {
			text = "???"
		}
		msg := tgbot.NewMessage(chatId, text)
		if _, err := to.bot.Send(msg); err != nil {
			return err
		}
	}
	return nil
}
