package modules

import "sync"

type SafeChan struct {
	C    chan struct{}
	once sync.Once
}

func NewSafeChan() *SafeChan {
	return &SafeChan{C: make(chan struct{}, 1)}
}

<<<<<<< HEAD
func (sc *SafeChan) Stop() {
=======
func (sc *SafeChan) SafeStop() {
>>>>>>> master-uibackend-0414
	sc.once.Do(func() {
		close(sc.C)
	})
}

