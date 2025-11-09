package telegram

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/hash"
	"strconv"
)

type RequestFromButton struct {
	RequestMessage string
	SecondAction   Action
}

type UsefulContentButtons map[string]RequestFromButton

var (
	globalUsefulContentButtons UsefulContentButtons
)

func init() {
	globalUsefulContentButtons = make(UsefulContentButtons)
}

// Создание кнопки с отсылаемым текстом и дополнительным действием
func MakeButton(text string, request string, action Action) tgbotapi.InlineKeyboardButton {
	str := text + request + strconv.Itoa(len(request)+len(text))
	globalUsefulContentButtons[hash.MD5(str)] =
		RequestFromButton{RequestMessage: request, SecondAction: action}
	return tgbotapi.NewInlineKeyboardButtonData(text, hash.MD5(str))
}

func GetGlobalUsefulContentButtons() UsefulContentButtons {
	return globalUsefulContentButtons
}
