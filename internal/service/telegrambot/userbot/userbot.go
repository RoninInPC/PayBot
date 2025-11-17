package userbot

import (
	"main/internal/database/queue"
	"main/internal/database/repository/factory"
	"main/internal/entity"
	"main/internal/service/telegrambot"
	"main/internal/telegram"
)

type UserBot struct {
	queueFromAdmin queue.Queue[entity.MessageFromAdminBot]
	queueFromUser  queue.Queue[entity.MessageFromUserBot]
	factory        factory.UnitOfWorkFactory
	telegrambot.TelegramBot
}

func InitUserBot(token string, queueFromAdmin queue.Queue[entity.MessageFromAdminBot], queueFromUser queue.Queue[entity.MessageFromUserBot], factory factory.UnitOfWorkFactory) UserBot {
	bot, err := telegrambot.InitBot(token)
	if err != nil {
		panic(err)
	}
	bot.TelegramCommands = bot.AddCommand(telegram.MakeUserRequestConfirmed(factory))
	return UserBot{TelegramBot: *bot, queueFromAdmin: queueFromAdmin, queueFromUser: queueFromUser, factory: factory}
}
