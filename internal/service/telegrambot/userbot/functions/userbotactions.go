package functions

import (
	"context"
	"github.com/and3rson/telemux/v2"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"log"
	"main/internal/database/repository/factory"
	"main/internal/model"
	"strings"
)

func StartUserBot(fac factory.UnitOfWorkFactory, u *telemux.Update) {
	if fac == nil {
		return
	}
	textParts := strings.Split(u.Update.Message.Text, " ")
	textPromocodeCode := ""
	if len(textParts) == 2 {
		textPromocodeCode = textParts[1]
	}

	user := u.Update.Message.From
	err := fac.New(context.Background(), pgx.Serializable, func(uow factory.UnitOfWork) error {
		promocodes, err := uow.PromocodeRepo().SelectByCode(context.Background(), []string{textPromocodeCode})
		if err != nil {
			return errors.Wrap(err, `select promocode (StartUserBot):`)
		}
		var promocodeId int64 = 0
		if len(promocodes) != 0 {
			promocodeId = promocodes[0].Id
		}

		users, err := uow.UserRepo().SelectByTgID(context.Background(), []int64{user.ID})
		if err != nil {
			return errors.Wrap(err, `select user (StartUserBot):`)
		}
		insertUser := model.User{TgID: }
		if len(users)!=0{
			insertUser = users[0]
		}
		uow.UserRepo().Upsert(context.Background(), []model.User{model.User{}})

		return nil
	})

	if err != nil {
		log.Println(err)
	}
}
