package timer

import (
	"time"
)

const (
	jiffies  time.Duration = 10 * time.Millisecond
	tvr_bits uint64        = 8
	tvn_bits uint64        = 6
	tvr_size uint64        = 1 << 8
	tvn_size uint64        = 1 << 6
	tvr_mask uint64        = tvr_size - 1
	tvn_mask uint64        = tvn_size - 1
)

// something like ThreadPoolExecutor in Java. See tools/gopool/SimpleGoPool.go
type Executor interface {
	Execute(action func()) error
}

type Timer struct {
	tv            []*wheel
	timer_jiffies int64 // 基准时间
	ticker        *time.Ticker
	stopChan      chan struct{}
	executor      Executor
}

func NewTimer(executor Executor) *Timer {
	t := &Timer{
		tv:            make([]*wheel, 5),
		stopChan:      make(chan struct{}),
		timer_jiffies: time.Now().UnixNano() / int64(jiffies),
		executor:      executor,
	}

	for i := 0; i < 5; i++ {
		if i == 0 {
			t.tv[0] = newWheel(int(tvr_size))
		} else {
			t.tv[i] = newWheel(int(tvn_size))
		}
	}
	return t
}

func (t *Timer) addTimertask(task *TimerTask) {
	expires := uint64(task.expires)
	idx := expires - uint64(t.timer_jiffies)

	if idx < tvr_size {
		slot := int(expires & tvr_mask)
		t.tv[0].addTask(slot, task)
	} else if idx < (1 << (tvr_bits + tvn_bits)) {
		slot := int((expires >> tvr_bits) & tvn_mask)
		t.tv[1].addTask(slot, task)
	} else if idx < (1 << (tvr_bits + 2*tvn_bits)) {
		slot := int((expires >> (tvr_bits + tvn_bits)) & tvn_mask)
		t.tv[2].addTask(slot, task)
	} else if idx < (1 << (tvr_bits + 3*tvn_bits)) {
		slot := int((expires >> (tvr_bits + 2*tvn_bits)) & tvn_mask)
		t.tv[3].addTask(slot, task)
	} else if int64(idx) < 0 {
		slot := int(uint64(t.timer_jiffies) & tvr_mask)
		t.tv[0].addTask(slot, task)
	} else {
		if idx > 0x00000000ffffffff {
			idx = 0x00000000fffffffff
			expires = idx + uint64(t.timer_jiffies)
		}
		slot := int((expires >> (tvr_bits + 3*tvn_bits)) & tvn_mask)
		t.tv[4].addTask(slot, task)
	}
}

// 计算任务过期时的jeffies值
func calcExpires(expireTime time.Time) time.Duration {
	return time.Duration(expireTime.UnixNano()) / jiffies
}

// 在firstTime运行后，然后每隔period运行一次
// 注：如果period>0,任务以固定频率重复执行，例如一个每分钟执行的task，
// 如果系统时间调大1小时，轮到它执行时，将一次性重复执行60次
// 如果period<0,任务以固定延迟重复执行：nextExecTime = curExecTime + |period|
func (t *Timer) RunAtFixedRate(name string, f func(), firstTime time.Time, period time.Duration) *TimerTask {
	task := newTimerTask(name, f, calcExpires(firstTime), period)
	t.addTimertask(task)
	return task
}

func (t *Timer) RunAtFixedDelay(name string, f func(), firstTime time.Time, period time.Duration) *TimerTask {
	return t.RunAtFixedRate(name, f, firstTime, -period)
}

func (t *Timer) RunWeeklyAt(name string, f func(), wkday time.Weekday, hour, minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	d += int(wkday-now.Weekday()+7) % 7
	runTime := time.Date(y, m, d, hour, minute, second, 0, now.Location())
	if now.After(runTime) {
		runTime = runTime.AddDate(0, 0, 7)
	}
	period := time.Hour * 24 * 7
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 指定时分秒，每天运行一次
func (t *Timer) RunDailyAt(name string, f func(), hour, minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, hour, minute, second, 0, now.Location())
	if now.After(runTime) {
		runTime = runTime.AddDate(0, 0, 1)
	}
	period := time.Hour * 24
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 指定分秒，每小时运行一次
func (t *Timer) RunHourlyAt(name string, f func(), minute, second int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), minute, second, 0, now.Location())
	period := time.Hour * 1
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 每N分钟运行一次
func (t *Timer) RunEveryNMinutes(name string, f func(), minutes int) *TimerTask {
	now := time.Now()
	y, m, d := now.Date()
	runTime := time.Date(y, m, d, now.Hour(), now.Minute()+minutes, 0, 0, now.Location()) // 首次运行延迟0~60s不等
	period := time.Minute * time.Duration(minutes)
	return t.RunAtFixedRate(name, f, runTime, period)
}

// 在指定时间运行一次
func (t *Timer) RunOnceAt(name string, f func(), runTime time.Time) *TimerTask {
	return t.RunAtFixedRate(name, f, runTime, 0)
}

func (t *Timer) Start() {
	go t.loop()
}

func (t *Timer) loop() {
	t.ticker = time.NewTicker(jiffies)
	for {
		select {
		case now, ok := <-t.ticker.C:
			if !ok { // closed
				break
			}
			t.handleTick(now)
		}
	}
}

// 获取tvn上的当前槽位，n为0时获取的是t.tv[1]
func (t *Timer) getTvnSlot(n int) int {
	s := (uint64(t.timer_jiffies) >> (tvr_bits + uint64(n)*tvn_bits)) & tvn_mask
	return int(s)
}

// 将wi所指的wheel上slot处的list中的任务重新加入timer
func (t *Timer) cascade(wi, slot int) bool {
	l := t.tv[wi].clearSlot(slot)
	if l.Len() == 0 {
		return slot == 0
	}
	for e := l.Front(); e != nil; e = e.Next() {
		task := e.Value.(*TimerTask)
		t.addTimertask(task)
	}
	return slot == 0
}

func (t *Timer) handleTick(now time.Time) {
	nowJiffies := now.UnixNano() / int64(jiffies)
	for nowJiffies >= t.timer_jiffies {
		idx := t.timer_jiffies & int64(tvr_mask)
		if idx == 0 &&
			t.cascade(1, t.getTvnSlot(0)) &&
			t.cascade(2, t.getTvnSlot(1)) &&
			t.cascade(3, t.getTvnSlot(2)) {
			t.cascade(4, t.getTvnSlot(3))
		}

		t.timer_jiffies += 1

		l := t.tv[0].clearSlot(int(idx))
		if l.Len() == 0 {
			continue
		}
		for e := l.Front(); e != nil; e = e.Next() {
			task := e.Value.(*TimerTask)
			f := task.getF()
			if t.executor != nil {
				t.executor.Execute(f)
			} else {
				f()
			}
			if task.period != 0 && !task.cancelled {
				if task.period > 0 {
					task.expires += (task.period / jiffies)
				} else {
					task.expires = time.Duration(nowJiffies) - (task.period / jiffies)
				}
				t.addTimertask(task)
			}
		}
	}
}

func (t *Timer) Stop() {
	t.ticker.Stop() // after this, handleTick may still running for a while
}
