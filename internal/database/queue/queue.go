package queue

type Queue[Anything any] interface {
	RPush(value Anything) error
	LPop() (Anything, error)
}
