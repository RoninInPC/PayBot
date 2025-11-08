package telegramentity

import (
	tm "github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func GetUserFromId(u *tm.Update) int64 {
	if u.Message != nil {
		return u.Message.From.ID
	} else if u.CallbackQuery != nil {
		return u.CallbackQuery.From.ID
	}
	return 0
}

func GetMessage(u *tm.Update) *tgbotapi.Message {
	if u.Message != nil {
		return u.Message
	} else if u.CallbackQuery != nil {
		return u.CallbackQuery.Message
	}
	return nil
}

func DeleteMessage(u *tm.Update) {
	_, _ = u.Bot.Send(tgbotapi.NewDeleteMessage(
		GetUserFromId(u),
		GetMessage(u).MessageID))
}
