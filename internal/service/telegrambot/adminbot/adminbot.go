package adminbot

import (
	"main/internal/database/cache"
	"main/internal/database/queue"
	"main/internal/database/repository/factory"
	"main/internal/entity"
	"main/internal/service/telegrambot"
	"main/internal/telegram"
)

type AdminBot struct {
	queueFromAdmin queue.Queue[entity.MessageFromAdminBot]
	queueFromUser  queue.Queue[entity.MessageFromUserBot]
	factory        factory.UnitOfWorkFactory
	state          telegram.StateHem
	telegrambot.TelegramBot
}

func InitAdminBot(token string, queueFromAdmin queue.Queue[entity.MessageFromAdminBot], queueFromUser queue.Queue[entity.MessageFromUserBot], factory factory.UnitOfWorkFactory, sets cache.Sets) AdminBot {
	bot, err := telegrambot.InitBot(token)
	if err != nil {
		panic(err)
	}
	state := telegram.InitState(sets)
	bot.TelegramCommands = bot.AddCommand(telegram.MakeUserRequestConfirmed(factory))
	return AdminBot{TelegramBot: *bot, queueFromAdmin: queueFromAdmin, queueFromUser: queueFromUser, factory: factory, state: state}
}
