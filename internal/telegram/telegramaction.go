package telegram

import (
	"github.com/and3rson/telemux/v2"
	"main/internal/database/entitybase"
	"main/internal/entity"
)

type Action interface {
	Action(u *telemux.Update)
}

type SimpleAction func(u *telemux.Update)

type SimpleActionStruct struct {
	SimpleAction SimpleAction
}

func (s SimpleActionStruct) Action(u *telemux.Update) {
	s.SimpleAction(u)
}

type UserCheckAction func(base entitybase.EntityBase[entity.User], u *telemux.Update)

type UserCheckActionStruct struct {
	Base         entitybase.EntityBase[entity.User]
	SimpleAction UserCheckAction
}

func (s UserCheckActionStruct) Action(u *telemux.Update) {
	s.SimpleAction(s.Base, u)
}
