package telegram

import (
	telemux "github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/database/entitybase"
	"main/internal/entity"
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
		"Analyser",
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

func MakeUserRequestConfirmed(base entitybase.EntityBase[entity.User]) TelegramCommand {
	return TelegramCommand{
		"Request",
		"",
		func(u *telemux.Update) bool {
			return u.ChatJoinRequest != nil
		},
		UserCheckActionStruct{
			Base: base,
			SimpleAction: func(base entitybase.EntityBase[entity.User], u *telemux.Update) {
				if base != nil {
					u.Bot.Send(tgbotapi.DeclineChatJoinRequest{
						ChatConfig: tgbotapi.ChatConfig{ChatID: u.ChatJoinRequest.Chat.ID},
						UserID:     u.ChatJoinRequest.From.ID,
					})
					return
				}
				user, err := base.Get(entity.User{UserName: u.ChatJoinRequest.From.UserName})
				if err != nil || user.ID == 0 {
					u.Bot.Send(tgbotapi.DeclineChatJoinRequest{
						ChatConfig: tgbotapi.ChatConfig{ChatID: u.ChatJoinRequest.Chat.ID},
						UserID:     u.ChatJoinRequest.From.ID,
					})
					return
				}

				u.Bot.Send(tgbotapi.ApproveChatJoinRequestConfig{
					ChatConfig: tgbotapi.ChatConfig{ChatID: u.ChatJoinRequest.Chat.ID},
					UserID:     u.ChatJoinRequest.From.ID,
				})
			},
		},
	}
}
