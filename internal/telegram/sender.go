package telegram

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type Sender interface {
	Send(tgbotapi.Chattable)
}

type TelegramSender struct {
	channel chan tgbotapi.Chattable
}

func InitTelegramSender(channel chan tgbotapi.Chattable) *TelegramSender {
	return &TelegramSender{
		channel: channel,
	}
}

func (telegramSender *TelegramSender) Send(chattable tgbotapi.Chattable) {
	telegramSender.channel <- chattable
}
