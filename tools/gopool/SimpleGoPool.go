package gopool

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
)

var ErrSimpleExecutorStopped = errors.New("SimpleExecutor stopped")

type Executor interface {
	Size() int
	Shutdown()
	Execute(command func()) error
}

type SimpleExecutor struct {
	stopped bool
	index   int
	size    int
	works   chan func()
	mu      sync.RWMutex
}

func NewSimpleExecutor(size int, queueSize int) Executor {
	if min := runtime.NumCPU()*2 + 1; size < min {
		size = min
	}
	e := &SimpleExecutor{
		index: 0,
		size:  size,
		works: make(chan func(), queueSize),
	}

	for i := 0; i < size; i++ {
		go func(i int) {
			for !e.stopped {
				select {
				case work := <-e.works:
					work()
				}
			}
			// if i == 0 {
			// 	close(e.works)
			// }
			// 执行已提交的任务
			for work := range e.works {
				work()
			}
		}(i)
	}
	return e
}

func (p *SimpleExecutor) Execute(f func()) error {
	safeF := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("Err:", err)
			}
		}()
		f()
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	if p.stopped {
		return ErrSimpleExecutorStopped
	}
	p.works <- safeF
	return nil
}

func (p *SimpleExecutor) Shutdown() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.stopped {
		return
	}
	p.stopped = true
	close(e.works)
}

func (p *SimpleExecutor) Size() int {
	return p.size
}
