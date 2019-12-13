package action

import (
	"fmt"
	"gamego/tools/gopool"
	"sync"
	"testing"
	"time"
)

type ExampleAction struct {
	DefalutAction
	Name string
	wg   *sync.WaitGroup
}

func (ea *ExampleAction) Execute() {
	fmt.Println(ea.Name, " ExampleAction-Execute")
	ea.wg.Done()
}

func TestAction(t *testing.T) {

	aq := NewActionQueue0(gopool.NewSimpleExecutor(2, 10))

	var wg sync.WaitGroup

	for i := 1; i < 11; i++ {
		wg.Add(1)
		action := &ExampleAction{
			DefalutAction: DefalutAction{aq},
			Name:          fmt.Sprintf("Action-%d", i),
			wg:            &wg,
		}
		aq.EnqueueAction(action)
	}
	wg.Wait()
}

type ExampleDelayAction struct {
	DefaultDelayAction
	Name string
	wg   *sync.WaitGroup
}

func (ea *ExampleDelayAction) Execute() {
	fmt.Println(ea.Name, " ExampleDelayAction-Execute")
	ea.wg.Done()
}

func TestDelayAction(t *testing.T) {
	aq := NewActionQueue0(gopool.NewSimpleExecutor(10, 10))

	var wg sync.WaitGroup

	begin := time.Now()
	for i := 1; i < 11; i++ {
		wg.Add(1)
		action := &ExampleDelayAction{
			DefaultDelayAction: DefaultDelayAction{
				Delay: 5 * int64(time.Second),
				Queue: aq},
			Name: fmt.Sprintf("DelayAction-%d", i),
			wg:   &wg,
		}
		aq.EnqueueDelayAction(action)
	}
	wg.Wait()
	end := time.Now()
	fmt.Println("Pass Seconds:", end.Unix()-begin.Unix())
}
