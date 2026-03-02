package functions

import (
	"context"
	"fmt"
	"log"
	database "main/internal/database/repository/factory"
	"main/internal/model"
	"main/internal/telegram"
	"slices"
	"strings"

	"github.com/and3rson/telemux/v2"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
)

func StartUserBot(sender telegram.Sender, fac database.UnitOfWorkFactory, u *telemux.Update, admins []int64) {
	if fac == nil {
		return
	}
	if slices.Contains(admins, u.Update.Message.From.ID) {
		msg := tgbotapi.NewMessage(u.Update.Message.Chat.ID, "Вы являетесь админом данного бота")
		sender.Send(msg)
		return
	}

	textParts := strings.Split(u.Update.Message.Text, " ")
	textPromocodeCode := ""
	if len(textParts) == 2 {
		textPromocodeCode = textParts[1]
	}

	user := u.Update.Message.From
	err := fac.New(context.Background(), pgx.Serializable, func(uow database.UnitOfWork) error {
		promocodes, err := uow.PromocodeRepo().SelectByCode(context.Background(), []string{textPromocodeCode})
		if err != nil {
			return errors.Wrap(err, `select promocode (StartUserBot):`)
		}
		var promocodeId int64 = 0
		containsProm := len(promocodes) != 0
		if containsProm {
			promocodeId = promocodes[0].Id
		}

		users, err := uow.UserRepo().SelectByTgID(context.Background(), []int64{user.ID})
		if err != nil {
			return errors.Wrap(err, `select user (StartUserBot):`)
		}
		insertUser := model.User{}
		containsUser := len(users) != 0
		if containsUser {
			insertUser = users[0]
		}
		insertUser.PromocodeID = &promocodeId
		_, err = uow.UserRepo().Upsert(context.Background(), []model.User{insertUser})
		if err == nil {
			text := ""
			if containsUser && containsProm {
				text = fmt.Sprintf("Пользователь %s воспользовался промокодом '%s'",*insertUser.Username,promocodes[0].Code)
			}
			if
			for _, admin := range admins {
				msg := tgbotapi.NewMessage(admin, text)
				sender.Send(msg)
			}
		} else {
			return errors.Wrap(err, `upsert user (StartUserBot):`)
		}
		return nil
	})

	if err != nil {
		log.Println(err)
	}
}
