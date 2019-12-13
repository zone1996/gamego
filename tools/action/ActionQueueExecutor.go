package action

import (
	"container/list"
	"gamego/tools/gopool"
	"sync"
	"time"
)

type ActionQueueExecutor struct {
	sync.Mutex
	executor     gopool.Executor
	delayActions *list.List
	ticker       *time.Ticker
}

func NewActionExexutor(executor gopool.Executor) *ActionQueueExecutor {
	aqe := &ActionQueueExecutor{
		executor:     executor,
		delayActions: list.New(),
		ticker:       time.NewTicker(10 * time.Millisecond),
	}
	go func() {
		for {
			select {
			case <-aqe.ticker.C:
				aqe.check()
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

func (aqe *ActionQueueExecutor) check() {
	aqe.Lock()
	defer aqe.Unlock()
	for e := aqe.delayActions.Front(); e != nil; {
		da := e.Value.(DelayAction)
		da.SetDelay(da.GetDelay() - int64(time.Millisecond)*10)
		if da.GetDelay() <= 0 {
			aqe.delayActions.Remove(e)
			da.GetQueue().EnqueueAction(da)
			e = aqe.delayActions.Front()
		} else {
			e = e.Next()
		}
	}
}
