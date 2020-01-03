package timer

import (
	"fmt"
	"time"
)

type TimerTask struct {
	f         func()
	name      string
	cancelled bool
	expires   time.Duration // 任务过期所需的jeffies数
	period    time.Duration // 至少一个jiffies，即10ms
}

func newTimerTask(name string, f func(), expires, period time.Duration) *TimerTask {
	if period > 0 && period < jiffies {
		period = 1 * jiffies
	} else if period < 0 && period > -jiffies {
		period = -1 * jiffies
	}
	return &TimerTask{f, name, false, expires, period}
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
