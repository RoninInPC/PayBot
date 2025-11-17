package telegram

import (
	"github.com/and3rson/telemux/v2"
	"main/internal/database/repository/factory"
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

type UserCheckAction func(factory factory.UnitOfWorkFactory, u *telemux.Update)

type UserCheckActionStruct struct {
	Factory      factory.UnitOfWorkFactory
	SimpleAction UserCheckAction
}

func (s UserCheckActionStruct) Action(u *telemux.Update) {
	s.SimpleAction(s.Factory, u)
}
