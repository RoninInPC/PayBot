package adminbot

import (
	"main/internal/database/queue"
	"main/internal/entity"
	"main/internal/service/telegrambot"
)

type AdminBot struct {
	queueFromAdmin queue.Queue[entity.MessageFromAdminBot]
	queueFromUser  queue.Queue[entity.MessageFromUserBot]
	telegrambot.TelegramBot
}
