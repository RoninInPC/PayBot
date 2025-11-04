package telegramentity

import (
	telemux "github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"strings"
)

type TelegramCommand struct {
	Name        string
	Description string
	Filter      telemux.FilterFunc
	Action      Action
}

type TelegramCommands []TelegramCommand

func (t TelegramCommands) AddCommand(command TelegramCommand) TelegramCommands {
	if len(t) == 0 {
		return TelegramCommands{command}
	}
	return append(t, command)
}

func FilterDefault(u *telemux.Update, name string) bool {
	if u.Message != nil {
		if strings.HasPrefix(u.Message.Text, "/"+name) {
			return true
		}
	}
	return false
}

func MakeCommandByFilterDefault(name, description string, action Action) TelegramCommand {
	return TelegramCommand{
		Name:        name,
		Description: description,
		Filter: func(u *telemux.Update) bool {
			return FilterDefault(u, name)
		},
		Action: action,
	}
}

func MakeFullCommand(name string, description string, filter telemux.FilterFunc, action Action) TelegramCommand {
	return TelegramCommand{
		Name:        name,
		Description: description,
		Filter:      filter,
		Action:      action,
	}
}

func MakeButtonAnalyser() TelegramCommand {
	return TelegramCommand{
		"ButtonAnalyser",
		"",
		func(u *telemux.Update) bool {
			return u.CallbackQuery != nil
		},
		SimpleActionStruct{
			SimpleAction: func(u *telemux.Update) {
				if u.CallbackQuery != nil {
					val, ok := GetGlobalUsefulContentButtons()[u.CallbackQuery.Data]
					if ok {
						if len(val.RequestMessage) > 0 {
							msg := tgbotapi.NewMessage(
								u.CallbackQuery.Message.Chat.ID,
								val.RequestMessage)
							_, _ = u.Bot.Send(msg)
						}
						if val.SecondAction != nil {
							val.SecondAction.Action(u)
							return
						}
					}
				}
			},
		},
	}
}
