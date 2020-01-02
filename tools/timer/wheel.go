package timer

import (
	"container/list"
)

type wheel struct {
	size  int
	slots []*list.List
}

func newWheel(size int) *wheel {
	w := &wheel{
		size:  size,
		slots: make([]*list.List, size),
	}
	for i := 0; i < w.size; i++ {
		w.slots[i] = list.New()
	}
	return w
}

func (w wheel) addTask(slot int, task *TimerTask) {
	if slot >= 0 && slot < w.size {
		w.slots[slot].PushBack(task)
	} else {
		panic("invalid slot for array")
	}
}
