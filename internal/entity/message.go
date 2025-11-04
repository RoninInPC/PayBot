package entity

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

type MessageFromAdminBot struct {
	TelegramID int
	tgbotapi.Message
}

type MessageFromUserBot struct {
	RequisiteContent []byte
	IsFile           bool
	IsImage          bool
	TariffPicked     Tariff
	PromoCodePicked  PromoCode
}
