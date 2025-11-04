package userbot

import (
	"main/internal/database/queue"
	"main/internal/entity"
	"main/internal/service/telegrambot"
)

type UserBot struct {
	queueFromAdmin queue.Queue[entity.MessageFromAdminBot]
	queueFromUser  queue.Queue[entity.MessageFromUserBot]
	telegrambot.TelegramBot
}
