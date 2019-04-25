package util

import (
	"sync"
)

// 自定义waitGroup，为了并发
type WaitGroupWrapper struct {
	sync.WaitGroup
}

// 异步执行，并w.add(1)
func (w *WaitGroupWrapper) Wrap(cb func()) {
	w.Add(1)
	go func() {
		cb()
		w.Done()
	}()
}
