package action

import (
	"container/list"
	"sync"
)

// 顺序执行入队的action
type ActionQueue struct {
	sync.Mutex
	actions  *list.List
	executor *ActionQueueExecutor
}

// 使用此方法时，ActionQueue独占一个goroutine扫描延迟任务
func NewActionQueue0(executor Executor) *ActionQueue {
	aq := &ActionQueue{
		actions:  list.New(),
		executor: NewActionExexutor(executor),
	}
	return aq
}

// 使用此方法时，共享ActionQueueExecutor的ActionQueue共用一个goroutine扫描延迟任务
func NewActionQueue1(executor *ActionQueueExecutor) *ActionQueue {
	aq := &ActionQueue{
		actions:  list.New(),
		executor: executor,
	}
	return aq
}

func (aq *ActionQueue) EnqueueAction(action Action) {
	aq.Lock()
	defer aq.Unlock()
	aq.actions.PushBack(action)
	if aq.actions.Len() > 1 {
		return
	}

	e := aq.actions.Front()
	f := func() {
		a := e.Value.(Action)
		a.Execute()
		aq.execNext(e)
	}
	aq.executor.Execute(f)
}

func (aq *ActionQueue) execNext(front *list.Element) {
	aq.Lock()
	defer aq.Unlock()
	aq.actions.Remove(front)
	if aq.actions.Len() < 1 {
		return
	}
	e := aq.actions.Front()
	f := func() {
		a := e.Value.(Action)
		a.Execute()
		aq.execNext(e)
	}
	aq.executor.Execute(f)
}

func (aq *ActionQueue) EnqueueDelayAction(delayAction DelayAction) {
	aq.executor.enQueueDelayAction(delayAction)
}
