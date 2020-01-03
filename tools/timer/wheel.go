package timer

import (
	"container/list"
	"sync"
)

type wheel struct {
	size  int
	slots []*list.List
	mu    sync.Mutex
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

func (w *wheel) clearSlot(slot int) *list.List {
	w.mu.Lock()
	defer w.mu.Unlock()
	l := w.slots[slot]
	if l.Len() != 0 {
		w.slots[slot] = list.New()
	}
	return l
}

func (w *wheel) addTask(slot int, task *TimerTask) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if slot >= 0 && slot < w.size {
		w.slots[slot].PushBack(task)
	} else {
		panic("invalid slot for array")
	}
}
