package telegrambot

import (
	"main/internal/telegram"
	"time"

	"github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramBot struct {
	telegram.Goroutines
	telegram.TelegramCommands
	admins  []int64
	channel chan tgbotapi.Chattable
	bot     *tgbotapi.BotAPI
}

func InitBot(token string, admins []int64) (*TelegramBot, error) {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TelegramBot{
		Goroutines: *telegram.InitGoroutines(),
		TelegramCommands: telegram.TelegramCommands{
			telegram.MakeButtonAnalyser(),
		},
		channel: make(chan tgbotapi.Chattable, 1000),
		admins:  admins,
		bot:     api}, nil
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

func (telegramBot *TelegramBot) Sender() telegram.Sender {
	return telegram.InitTelegramSender(telegramBot.channel)
}

func (telegramBot *TelegramBot) Work() {
	telegramBot.initBotMenu()
	go func() {
		for chattable := range telegramBot.channel {
			_, err := telegramBot.bot.Send(chattable)
			if err != nil {
				telegramBot.channel <- chattable
				time.Sleep(time.Millisecond * 10)
			}
		}
	}()
	telegramBot.dispatchUpdates()
}
