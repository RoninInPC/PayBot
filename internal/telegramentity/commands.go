package telegramentity

import (
	telemux "github.com/and3rson/telemux/v2"
	"strings"
)

type TelegramCommand struct {
	Name        string
	Description string
	Filter      telemux.FilterFunc
	Action      Action
}

type TelegramCommands []TelegramCommand

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
