package action

import (
	"container/list"
	"sync"
	"time"
)

// something like ThreadPoolExecutor in Java. See tools/gopool/SimpleGoPool.go
type Executor interface {
	Execute(action func()) error
}

type ActionQueueExecutor struct {
	sync.Mutex
	executor     Executor
	delayActions *list.List
	ticker       *time.Ticker
	step         int64
}

func NewActionExexutor(executor Executor) *ActionQueueExecutor {
	step := 100 * int64(time.Millisecond) // 100ms 扫描一次
	aqe := &ActionQueueExecutor{
		executor:     executor,
		delayActions: list.New(),
		ticker:       time.NewTicker(time.Duration(step)),
		step:         step,
	}
	go func() {
		for {
			select {
			case now, ok := <-aqe.ticker.C:
				if !ok {
					return
				}
				aqe.check(now.UnixNano())
			}
		}
	}()
	return aqe
}

func (aqe *ActionQueueExecutor) Execute(f func()) {
	aqe.executor.Execute(f)
}

func (aqe *ActionQueueExecutor) enQueueDelayAction(delayAction DelayAction) {
	aqe.Lock()
	defer aqe.Unlock()
	aqe.delayActions.PushBack(delayAction)
}

func (aqe *ActionQueueExecutor) check(now int64) {
	aqe.Lock()
	defer aqe.Unlock()
	for e := aqe.delayActions.Front(); e != nil; {
		da := e.Value.(DelayAction)
		if da.GetExecTime() <= now {
			aqe.delayActions.Remove(e)
			da.GetQueue().EnqueueAction(da)
			e = aqe.delayActions.Front()
		} else {
			e = e.Next()
		}
	}
}
