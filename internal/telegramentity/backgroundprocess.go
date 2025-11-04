package telegramentity

type Running chan bool

type Goroutines struct {
	Container map[string]Running
}

func InitGoroutines() *Goroutines {
	return &Goroutines{
		Container: make(map[string]Running)}
}

func (goroutines *Goroutines) AddGlobalGoroutine(
	nameProcess string,
	anotherFunc Action) *Goroutines {
	_, ok := goroutines.Container[nameProcess]
	if ok {
		return goroutines
	}
	stg := make(Running)
	goroutines.Container[nameProcess] = stg
	go func() {
		for {
			select {
			case <-goroutines.Container[nameProcess]:
				return
			default:
				anotherFunc.Action(nil)
			}
		}
	}()
	return goroutines
}

func (goroutines *Goroutines) DeleteGoroutineByName(nameProcess string) *Goroutines {
	channel, ok := goroutines.Container[nameProcess]
	if !ok {
		return goroutines
	}
	channel <- true
	close(channel)
	delete(goroutines.Container, nameProcess)
	return goroutines
}
