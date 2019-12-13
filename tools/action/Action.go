package action

import (
	"fmt"
)

type Action interface {
	Execute()
	GetQueue() *ActionQueue
}

type DelayAction interface {
	Action
	GetDelay() int64 // 毫秒, 延迟时间=Delay()*time.Millisecond
	SetDelay(delay int64)
}

// 示例1
type DefalutAction struct {
	Queue *ActionQueue
}

func (da *DefalutAction) Execute() {
	fmt.Println("ExampleAction-Execute")
}

func (da *DefalutAction) GetQueue() *ActionQueue {
	return da.Queue
}

// 示例2
type DefaultDelayAction struct {
	Delay int64
	Queue *ActionQueue
}

func (dda *DefaultDelayAction) Execute() {
	fmt.Println("ExampleDelayAction-Execute")
}

func (dda *DefaultDelayAction) GetQueue() *ActionQueue {
	return dda.Queue
}

func (dda *DefaultDelayAction) GetDelay() int64 {
	return dda.Delay
}

func (dda *DefaultDelayAction) SetDelay(delay int64) {
	dda.Delay = delay
}
