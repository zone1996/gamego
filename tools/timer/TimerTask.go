package timer

import (
	"fmt"
)

type TimerTask struct {
	f         func()
	name      string
	cancelled bool
	turn      int64
	period    int64 // Nanosecond, 0:run once
}

func newTimerTask(name string, f func(), period int64) *TimerTask {
	return &TimerTask{
		f:      f,
		name:   name,
		period: period,
	}
}

func (t *TimerTask) getF() func() {
	f := func() {
		defer func() {
			if err := recover(); err != nil {
				fmt.Println("定时器任务:", t.name, ", err:", err)
			}
		}()
		if !t.cancelled {
			t.f()
		}
	}
	return f
}

func (t *TimerTask) Cancel() {
	t.cancelled = true
}
