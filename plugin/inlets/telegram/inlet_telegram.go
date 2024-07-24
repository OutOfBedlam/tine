package telegram

import (
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
	"sync"
	"time"

	tg "github.com/OutOfBedlam/tine/drivers/telegram"
	"github.com/OutOfBedlam/tine/engine"
	tgbot "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func init() {
	engine.RegisterInlet(&engine.InletReg{
		Name:    "telegram",
		Factory: TelegramInlet,
	})
}

func TelegramInlet(ctx *engine.Context) engine.Inlet {
	return &telegramInlet{ctx: ctx}
}

type telegramInlet struct {
	ctx        *engine.Context
	bot        *tgbot.BotAPI
	closeOnce  sync.Once
	httpClient *http.Client
}

var _ = engine.PushInlet((*telegramInlet)(nil))

func (ti *telegramInlet) Open() error {
	token := ti.ctx.Config().GetString("token", "")

	if bot, err := tg.GetOrNewBot(token); err != nil {
		return err
	} else {
		ti.bot = bot
		ti.bot.Debug = ti.ctx.Config().GetBool("debug", false)
	}
	return nil
}

func (ti *telegramInlet) Close() error {
	ti.closeOnce.Do(func() {
		if ti.bot != nil {
			ti.bot.StopReceivingUpdates()
		}
	})
	return nil
}

func (ti *telegramInlet) Push(cb func([]engine.Record, error)) {
	u := tgbot.NewUpdate(0)
	u.Timeout = int((ti.ctx.Config().GetDuration("timeout", 5*time.Second)).Seconds())
	updates := ti.bot.GetUpdatesChan(u)
	for update := range updates {
		if update.Message == nil {
			continue
		}
		rec := engine.NewRecord(
			engine.NewIntField("message_id", int64(update.Message.MessageID)),
			engine.NewStringField("from", update.Message.From.String()),
			engine.NewStringField("chat_title", update.Message.Chat.Title),
			engine.NewIntField("chat_id", int64(update.Message.Chat.ID)),
			engine.NewStringField("text", update.Message.Text),
		)
		if len(update.Message.Photo) > 0 {
			photo := update.Message.Photo[len(update.Message.Photo)-1]
			tfile, err := ti.bot.GetFile(tgbot.FileConfig{FileID: photo.FileID})
			if err != nil {
				ti.ctx.LogWarn("inetls.telegram get photo", "error", err)
				continue
			}
			photoUrl := tfile.Link(ti.bot.Token)
			if bin, _, err := ti.fetchHttp(photoUrl); err != nil {
				ti.ctx.LogWarn("inetls.telegram fetch photo", "error", err)
				continue
			} else {
				bv := engine.NewBinaryField("photo", bin)
				// since telegram server returns "application/octet-stream"
				// we need to set the content type manually
				switch strings.ToLower(filepath.Ext(tfile.FilePath)) {
				case ".png":
					bv.SetTag("Content-Type", "image/png")
				case ".jpg", ".jpeg":
					bv.SetTag("Content-Type", "image/jpeg")
				case ".gif":
					bv.SetTag("Content-Type", "image/gif")
				default:
					ti.ctx.LogWarn("inetls.telegram fetch photo unknown file type", "file", tfile.FilePath)
					continue
				}
				ti.ctx.LogDebug("inlets.telegram", "photo", tfile.FilePath, "size", photo.FileSize)
				rec = rec.Append(bv)
				if field := rec.Field("text"); field.Value.Raw().(string) == "" {
					if txt := update.Message.Caption; txt == "" {
						field.Value = engine.NewValue("Explain this photo.")
					} else {
						field.Value = engine.NewValue(txt)
					}
				}
			}
		}
		cb([]engine.Record{rec}, nil)
	}
	cb(nil, io.EOF)
}

func (ai *telegramInlet) fetchHttp(addr string) ([]byte, string, error) {
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		return nil, "", fmt.Errorf("unsupported protocol: %s", addr)
	}
	if ai.httpClient == nil {
		transport := &http.Transport{
			Proxy: http.ProxyFromEnvironment,
		}
		ai.httpClient = &http.Client{
			Transport: transport,
		}
	}
	rsp, err := ai.httpClient.Get(addr)
	if err != nil {
		return nil, "", err
	}
	defer rsp.Body.Close()
	if body, err := io.ReadAll(rsp.Body); err != nil {
		return nil, "", err
	} else {
		return body, rsp.Header.Get("Content-Type"), nil
	}
}
