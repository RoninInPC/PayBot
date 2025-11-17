package cache

type Sets interface {
	Add(name, value string) error
	GetAll(name string) ([]string, error)
	Delete(name, value string) error
	Clear(name string) error
}
