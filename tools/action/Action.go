package action

import (
	"time"
)

type Action interface {
	Execute()
	GetQueue() *ActionQueue
}

type DelayAction interface {
	Action
	GetExecTime() int64
	SetDelay(delay int64)
}

// adapter for Action interface
type defalutAction struct {
	Queue *ActionQueue
}

func (da *defalutAction) Execute() {
	panic("Please override this method")
}

func (da *defalutAction) GetQueue() *ActionQueue {
	return da.Queue
}

//  adapter for DelayAction interface
type defaultDelayAction struct {
	delay    int64
	execTime int64
	Queue    *ActionQueue
}

func (dda *defaultDelayAction) Execute() {
	panic("Please override this method")
}

func (dda *defaultDelayAction) GetQueue() *ActionQueue {
	return dda.Queue
}

func (dda *defaultDelayAction) SetDelay(delay int64) {
	dda.delay = delay
	dda.execTime = time.Now().UnixNano() + delay
}

func (dda *defaultDelayAction) GetExecTime() int64 {
	return dda.execTime
}

func NewAction(queue *ActionQueue) Action {
	return &defalutAction{
		Queue: queue,
	}
}

func NewDelayAction(delay int64, queue *ActionQueue) DelayAction {
	da := &defaultDelayAction{
		delay:    delay,
		Queue:    queue,
		execTime: time.Now().UnixNano() + delay,
	}
	return da
}
