package telegram

import (
	"encoding/json"
	"github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"main/internal/entity"
	"time"
)

func CreateInviteLinkToUser(resource *entity.Resource, user *entity.User, update telemux.Update, expiredAt time.Time) (link tgbotapi.ChatInviteLink, err error) {
	unbanConfig := tgbotapi.UnbanChatMemberConfig{ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: resource.ChatId, UserID: user.UserTelegramId}}
	update.Bot.Request(unbanConfig)

	inviteLinkConfig := tgbotapi.CreateChatInviteLinkConfig{ChatConfig: tgbotapi.ChatConfig{ChatID: resource.ChatId}, CreatesJoinRequest: true, ExpireDate: int(expiredAt.Unix())}
	resp, err := update.Bot.Request(inviteLinkConfig)
	if err != nil {
		return link, err
	}
	err = json.Unmarshal(resp.Result, &link)
	return link, err
}

func KickUserFromResource(resource *entity.Resource, user *entity.User, update telemux.Update) error {
	banConfig := tgbotapi.BanChatMemberConfig{ChatMemberConfig: tgbotapi.ChatMemberConfig{ChatID: resource.ChatId, UserID: user.UserTelegramId}}
	_, err := update.Bot.Request(banConfig)
	return err
}
