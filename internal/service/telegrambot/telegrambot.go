package telegrambot

import (
	"github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/telegramentity"
)

type TelegramBot struct {
	telegramentity.Goroutines
	telegramentity.TelegramCommands
	bot *tgbotapi.BotAPI
}

func InitBot(token string) (*TelegramBot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		Goroutines:       *telegramentity.InitGoroutines(),
		TelegramCommands: telegramentity.TelegramCommands{telegramentity.MakeButtonAnalyser()},
		bot:              api}, nil
}

func (telegramBot *TelegramBot) initBotMenu() {
	var sliceArr []tgbotapi.BotCommand
	for _, action := range telegramBot.TelegramCommands {
		if len(action.Description) > 0 {
			sliceArr = append(sliceArr, tgbotapi.BotCommand{
				Command:     action.Name,
				Description: action.Description,
			})
		}
	}
	cmdCfg := tgbotapi.NewSetMyCommands(
		sliceArr...,
	)
	_, _ = telegramBot.bot.Send(cmdCfg)
}

func (telegramBot *TelegramBot) getUpdates(timeOut int) tgbotapi.UpdatesChannel {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = timeOut
	return telegramBot.bot.GetUpdatesChan(u)
}

func (telegramBot *TelegramBot) dispatchUpdates() {
	mux := telemux.NewMux()
	for _, command := range telegramBot.TelegramCommands {
		mux.AddHandler(telemux.NewHandler(command.Filter, func(u *telemux.Update) {
			command.Action.Action(u)
		}))

	}
	for update := range telegramBot.getUpdates(40) {
		mux.Dispatch(telegramBot.bot, update)
	}
}

func (telegramBot *TelegramBot) Work() {
	telegramBot.initBotMenu()
	telegramBot.dispatchUpdates()
}
