package mapper

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/entity"
)

func SentMessageToSend(message entity.MessageFromAdminBot) tgbotapi.Chattable {
	id := message.TelegramID
}
