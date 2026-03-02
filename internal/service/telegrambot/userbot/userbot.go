package userbot

import (
	database "main/internal/database/repository/factory"
	"main/internal/service/telegrambot"
	"main/internal/service/telegrambot/userbot/functions"
	"main/internal/telegram"
)

type UserBot struct {
	factory database.UnitOfWorkFactory
	telegrambot.TelegramBot
}

func InitUserBot(token string, admins []int64, factoryOfUnits database.UnitOfWorkFactory) UserBot {
	bot, err := telegrambot.InitBot(token, admins)
	if err != nil {
		panic(err)
	}
	bot.TelegramCommands =
		bot.AddCommand(telegram.MakeUserRequestConfirmed(factoryOfUnits)).
			AddCommand(telegram.MakeCommandByFilterDefault(
				"start",
				"Начнём?",
				telegram.FactoryActionStruct{

					Admins:       admins,
					Factory:      factoryOfUnits,
					SimpleAction: functions.StartUserBot}))
	return UserBot{TelegramBot: *bot, factory: factoryOfUnits}
}
