package modules

import "sync"

type SafeChan struct {
	C    chan struct{}
	once sync.Once
}

func NewSafeChan() *SafeChan {
	return &SafeChan{C: make(chan struct{}, 1)}
}

func (sc *SafeChan) SafeStop() {
	sc.once.Do(func() {
		close(sc.C)
	})
}

