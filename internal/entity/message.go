package entity

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MessageFromAdminBot struct {
	TelegramID int
	tgbotapi.Message
}
