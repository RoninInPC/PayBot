package telegram

import (
	"main/internal/database/repository/factory"

	"github.com/and3rson/telemux/v2"
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

type FactoryAction func(sender Sender, factory database.UnitOfWorkFactory, u *telemux.Update, admins []int64)

type FactoryActionStruct struct {
	Sender       Sender
	Factory      database.UnitOfWorkFactory
	SimpleAction FactoryAction
	Admins       []int64
}

func (s FactoryActionStruct) Action(u *telemux.Update) {
	s.SimpleAction(s.Sender, s.Factory, u, s.Admins)
}
