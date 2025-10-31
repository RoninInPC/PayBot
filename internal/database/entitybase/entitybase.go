package entitybase

type EntityBase[Anything any] interface {
	Add(Anything) error
	Update(Anything) error
	Get(Anything) (Anything, error)
	Delete(Anything) error
	GetAll() ([]Anything, error)
}
