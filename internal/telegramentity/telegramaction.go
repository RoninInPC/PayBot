package telegramentity

import (
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
