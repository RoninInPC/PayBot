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
func MakeInlineButton(text string, request string, action Action) tgbotapi.InlineKeyboardButton {
	str := text + request + strconv.Itoa(len(request)+len(text))
	hashMD5 := hash.MD5(str)
	globalUsefulContentButtons[hashMD5] =
		RequestFromButton{RequestMessage: request, SecondAction: action}
	return tgbotapi.NewInlineKeyboardButtonData(text, hashMD5)
}

func MakeButtonWebApp(text string, url string) tgbotapi.KeyboardButton {
	return tgbotapi.NewKeyboardButtonWebApp(text, tgbotapi.WebAppInfo{URL: url})
}

func GetGlobalUsefulContentButtons() UsefulContentButtons {
	return globalUsefulContentButtons
}
