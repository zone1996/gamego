package event

import (
	"container/list"
	"reflect"
	"sync"
)

type Event struct {
	EventType int           // 事件类型
	EventData []interface{} // 事件携带的数据
}

func NewEvent(eventType int, data ...interface{}) *Event {
	e := &Event{
		EventType: eventType,
		EventData: data,
	}
	return e
}

// 事件监听器
type EventListener interface {
	OnEvent(event *Event)
}

// 事件接收器
type EventReceiver struct {
	mu     sync.Mutex
	events map[int]*list.List
}

func NewEventReceiver() *EventReceiver {
	return &EventReceiver{
		events: make(map[int]*list.List),
	}
}

func (er *EventReceiver) AddListener(listener EventListener, eventType int) {
	er.mu.Lock()
	defer er.mu.Unlock()
	listeners, ok := er.events[eventType]
	if !ok {
		listeners = list.New()
		er.events[eventType] = listeners
	}
	for e := listeners.Front(); e != nil; e = e.Next() {
		if reflect.DeepEqual(e.Value.(EventListener), listener) {
			return
		}
	}
	listeners.PushBack(listener)
}

func (er *EventReceiver) RemoveListener(listener EventListener, eventType int) {
	er.mu.Lock()
	defer er.mu.Unlock()
	listeners, ok := er.events[eventType]
	if !ok {
		return
	}
	for e := listeners.Front(); e != nil; e = e.Next() {
		if reflect.DeepEqual(e.Value.(EventListener), listener) {
			listeners.Remove(e)
			break
		}
	}
}

func (er *EventReceiver) RemoverListeners(eventType int) {
	er.mu.Lock()
	defer er.mu.Unlock()
	delete(er.events, eventType)
}

func (er *EventReceiver) ReceiveEvent(event *Event) {
	er.mu.Lock()
	defer er.mu.Unlock()
	listeners, ok := er.events[event.EventType]
	if !ok {
		return
	}
	// 一般一个事件对应一个EventListener, 挨个执行应该不会阻塞太久
	for e := listeners.Front(); e != nil; e = e.Next() {
		listener := e.Value.(EventListener)
		listener.OnEvent(event)
	}
}

func (er *EventReceiver) Clear() {
	er.mu.Lock()
	defer er.mu.Unlock()
	er.events = make(map[int]*list.List)
}
