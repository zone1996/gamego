package timer

import (
	"container/list"
	"gamego/tools/gopool"
	"time"
)

type Timer struct {
	curSlot      int64 // 0~totalSlotNum-1
	totalSlotNum int64
	step         int64 // Nanosecond
	slots        []*list.List
	ticker       *time.Ticker
	stopChan     chan struct{}
	executor     gopool.Executor
}

func NewDefaultTimer() *Timer {
	var slotNum int64 = 50
	t := &Timer{
		curSlot:      0,
		totalSlotNum: slotNum,
		step:         int64(time.Second) / slotNum, // 20ms
		slots:        make([]*list.List, slotNum),
		stopChan:     make(chan struct{}, 1),
		executor:     gopool.NewSimpleExecutor(20, 1000),
	}
	return t
}

// 在firstTime运行后，每隔period秒运行一次
func (t *Timer) RunAtFixedRate(name string, f func(), firstTime time.Time, period int64) *TimerTask {
	task := newTimerTask(name, f, period)
	turn, slot := t.calcTurnAndSlot(firstTime.UnixNano() - time.Now().UnixNano())
	task.turn = turn
	t.addTask(task, slot)
	return task
}

// 指定时分秒，每天运行一次
func (t *Timer) RunDailyAt(name string, f func(), hour, minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, hour, minute, second, 0, now.Location())
	if now.After(runTime) {
		runTime = runTime.AddDate(0, 0, 1)
	}
	period := int64(time.Hour) * 24
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 指定分秒，每小时运行一次
func (t *Timer) RunHourlyAt(name string, f func(), minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), minute, second, 0, now.Location())
	period := int64(time.Hour)
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 每N分钟运行一次
func (t *Timer) RunEveryNMinutes(name string, f func(), minutes int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), now.Minute()+minutes, 0, 0, now.Location()) // 首次运行延迟0~60s不等
	period := int64(time.Minute) * int64(minutes)
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 在指定时间运行一次
func (t *Timer) RunOnceAt(name string, f func(), runTime time.Time) *TimerTask {
	return t.RunAtFixedRate(name, f, runTime, 0)
}

func (t *Timer) addTask(task *TimerTask, slot int64) {
	if t.slots[slot] == nil {
		t.slots[slot] = list.New()
	}
	t.slots[slot].PushBack(task)
}

func (t *Timer) calcTurnAndSlot(delayTime int64) (int64, int64) {
	if delayTime <= 0 {
		return 0, 0
	}
	delayStep := delayTime / t.step
	turn := delayStep / t.totalSlotNum
	curSlot := t.curSlot

	residue := delayStep % t.totalSlotNum
	slot := (residue + curSlot) % t.totalSlotNum
	if residue+curSlot > t.totalSlotNum {
		turn += 1
	}
	return turn, slot
}

// 重新安排
func (t *Timer) reSchedule(task *TimerTask) {
	turn, slot := t.calcTurnAndSlot(task.period)
	task.turn = turn
	t.addTask(task, slot)
}

func (t *Timer) Start() {
	go t.start0()
}

func (t *Timer) start0() {
	t.ticker = time.NewTicker(time.Duration(t.step) * time.Nanosecond)
	lastTickTime := time.Now()
	for {
		select {
		case now, ok := <-t.ticker.C:
			if !ok { // closed
				t.slots = nil
				t.executor.Shutdown()
				break
			}
			if now.Unix()-lastTickTime.Unix() > 5 { // TODO 机器时间发生变化:两次tick的Unix时间间隔超过5s
			}
			lastTickTime = now
			t.handleTick()
		}
	}
}

func (t *Timer) handleTick() {
	defer func() {
		t.curSlot++
		t.curSlot = t.curSlot % t.totalSlotNum
	}()

	tasks := t.slots[t.curSlot]
	if tasks == nil {
		return
	}

	for e := tasks.Front(); e != nil; {
		task := e.Value.(*TimerTask)
		task.turn--
		if task.turn > 0 {
			e = e.Next()
			continue
		}
		f := task.getF()
		t.executor.Execute(f)
		temp := e.Next()
		tasks.Remove(e)
		e = temp
		if task.period != 0 && !task.cancelled {
			t.reSchedule(task) // 先remove后reschedule，避免刚加入就turn--
		}
	}
}

func (t *Timer) Stop() {
	t.ticker.Stop() // after this, handleTick may still running for a while
}

func (t *Timer) StopNow() {
	t.ticker.Stop()
	t.executor.Shutdown()
}
