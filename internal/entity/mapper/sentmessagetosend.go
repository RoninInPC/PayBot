package mapper

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/entity"
)

func FilesToInterface(files []tgbotapi.FilePath) []interface{} {
	result := make([]interface{}, len(files))
	for i, file := range files {
		result[i] = file
	}
	return result
}

func SentMessageToSend(message entity.MessageFromAdminBot) []tgbotapi.Chattable {
	filesInput := make([]interface{}, 0)
	answer := make([]tgbotapi.Chattable, 0)
	for _, file := range message.Files {
		filePath := tgbotapi.FilePath(file.Filename)
		switch file.Type {
		case entity.Doc:
			filesInput = append(filesInput, tgbotapi.NewInputMediaDocument(filePath))
		case entity.Photo:
			filesInput = append(filesInput, tgbotapi.NewInputMediaPhoto(filePath))
		case entity.Video:
			filesInput = append(filesInput, tgbotapi.NewInputMediaVideo(filePath))
		case entity.Audio:
			filesInput = append(filesInput, tgbotapi.NewInputMediaAudio(filePath))
		case entity.Animation:
			filesInput = append(filesInput, tgbotapi.NewInputMediaAnimation(filePath))
		case entity.Voice:
			answer = append(answer, tgbotapi.NewVoice(message.TelegramID, filePath))
			return answer
		case entity.VideoVoice:
			answer = append(answer, tgbotapi.NewVideoNote(message.TelegramID, 60, filePath))
			return answer
		}
	}
	if len(filesInput) > 0 {
		answer = append(answer, tgbotapi.NewMediaGroup(message.TelegramID, filesInput))
	}
	if len(message.Text) != 0 {
		answer = append(answer, tgbotapi.NewMessage(message.TelegramID, message.Text))
	}
	return answer
}
