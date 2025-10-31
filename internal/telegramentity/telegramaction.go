package telegramentity

import (
	"github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Action interface {
	Action(u *telemux.Update)
}

type SimpleAction func(telegramBot *tgbotapi.BotAPI, u *telemux.Update)

type SimpleActionStruct struct {
	SimpleAction SimpleAction
	TelegramBot  *tgbotapi.BotAPI
}

func (s SimpleActionStruct) Action(u *telemux.Update) {
	s.SimpleAction(s.TelegramBot, u)
}
