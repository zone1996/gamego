package timer

import (
	"container/list"
	"time"
)

type Timer struct {
	curSlot      int64 // 0~totalSlotNum-1
	totalSlotNum int64
	slots        []*list.List

	ticker   *time.Ticker
	step     time.Duration
	stopChan chan struct{}
}

func NewDefaultTimer() *Timer {
	t := &Timer{
		curSlot:      0,
		totalSlotNum: 50,
		slots:        make([]*list.List, 50),
		step:         time.Second / 50, // 20ms
		stopChan:     make(chan struct{}, 1),
	}
	return t
}

// 在firstTime运行后，每隔period秒运行一次
func (t *Timer) RunAtFixedRate(name string, f func(), firstTime time.Time, period int64) *TimerTask {
	task := newTimerTask(name, f, period)
	turn, slot := calcTurnAndSlot(firstTime)
	task.turn = turn
	t.addTask(task, slot)
	return task
}

// 在每天指定的时分秒运行一次
func (t *Timer) RunDailyAt(name string, f func(), hour, minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, hour, minute, second, 0, now.Location())
	period := int64(time.Hour) * 24
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 在指定的分秒，每小时运行一次
func (t *Timer) RunHourlyAt(name string, f func(), minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), minute, second, 0, now.Location())
	period := int64(time.Hour)
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 每分钟运行一次
func (t *Timer) RunEveryNMinutes(name string, f func(), minutes int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), now.Minute()+minutes, 0, 0, now.Location())
	period := int64(time.Minute) * int64(minutes)
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 在指定时间运行一次
func (t *Timer) RunOnceAt(name string, f func(), runTime time.Time) *TimerTask {
	return t.RunAtFixedRate(name, f, runTime, 0)
}

func (t *Timer) calcTurnAndSlot(runTime time.Time) (int64, int64) {
	delayStep := (runTime.UnixNano() - time.Now().UnixNano()) / int64(step)
	turn := delayStep / t.totalSlotNum
	curSlot := t.curSlot

	residue := delayStep % t.totalSlotNum
	if residue+curSlot < t.totalSlotNum {
		return turn, residue + curSlot
	} else {
		return turn + 1, residue + curSlot - t.totalSlotNum
	}
}

// 重新安排
func (t *Timer) reSchedule(task *TimerTask) {
	delayStep := task.period / int64(step)
	turn := delayStep / t.totalSlotNum
	curSlot := t.curSlot

	residue := delayStep % t.totalSlotNum
	slot := (residue + curSlot) % t.totalSlotNum
	if residue+curSlot > t.totalSlotNum {
		turn += 1
	}
	task.turn = turn
	addTask(task, slot)
}

func (t *Timer) addTask(task *TimerTask, slot int) {
	if t.slots[slot] == nil {
		t.slots[slot] = list.New()
	}
	t.slots[slot].PushBack(task)
}

func (t *Timer) Start() {
	go t.start0()
}

func (t *Timer) start0() {
	t.ticker = time.NewTicker(t.step)
	lastTickTime := time.Now()
	for {
		select {
		case <-t.ticker.C:
			now := time.Now()
			if now.Unix()-lastTickTime.Unix() > 5 { // 两次tick的Unix时间间隔超过5s
				// TODO 机器时间发生变化
			}
			lastTickTime = now
			t.handleTick()
		case <-t.stopChan:
			t.ticker.Stop()
			t.slots = nil
			close(t.stopChan)
			return
		}
	}
}

func (t *Timer) handleTick() {
	tasks := t.slots[t.curSlot]
	if tasks == nil {
		return
	}

	for e := tasks.Front(); e != nil; {
		task := e.(*TimerTask)
		if task.turn -= 1; task.turn > 0 {
			continue
		}
		go task.run()
		temp := e.Next()
		tasks.Remove(e)
		e = temp
		if task.period != 0 && !task.cancelled {
			t.reSchedule(task)
		}
	}
}

func (t *Timer) Stop() {
	t.stopChan <- struct{}{}
}
