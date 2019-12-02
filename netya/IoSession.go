package netya

import (
	"bytes"
	"net"
	"sync"
	"sync/atomic"

	log "github.com/zone1996/logo"
)

type IoSession struct {
	Id             int32
	conn           net.Conn
	InBoundBuffer  *bytes.Buffer // using for cumulate bytes
	OutBoundBuffer *bytes.Buffer // using for async write
	AsyncWriteChan chan struct{}
	AsyncTaskChan  chan func()
	Attribute      map[string]interface{}
	AliveState     int32 // 1:alive
	mu             sync.RWMutex
}

func NewIoSession(conn net.Conn) *IoSession {
	session := &IoSession{
		conn:           conn,
		InBoundBuffer:  new(bytes.Buffer),
		OutBoundBuffer: new(bytes.Buffer),
		Attribute:      make(map[string]interface{}),
		AsyncWriteChan: make(chan struct{}),
		AsyncTaskChan:  make(chan func(), 64),
		AliveState:     1,
	}
	return session
}

func (this *IoSession) SetId(id int32) {
	this.Id = id
}

func (this *IoSession) GetId() int32 {
	return this.Id
}

func (this *IoSession) IsAlive() bool {
	return atomic.LoadInt32(&this.AliveState) == 1
}

func (this *IoSession) SetAlive(alive bool) {
	var v int32 = 0
	if alive {
		v = 1
	}
	atomic.SwapInt32(&this.AliveState, v)
}

func (this *IoSession) Write(b []byte) (n int, err error) {
	if this.IsAlive() {
		return this.conn.Write(b)
	}
	return 0, nil
}

func (this *IoSession) AsyncWrite(b []byte) {
	if this.IsAlive() {
		this.mu.Lock()
		defer this.mu.Unlock()
		n, err := this.OutBoundBuffer.Write(b)
		if err != nil {
			log.Info("AsyncWrite err: ?", err)
		}
		if n > 0 {
			this.AsyncWriteChan <- struct{}{}
		}
	}
}

func (this *IoSession) doAsyncWrite() {
	for _ = range this.AsyncWriteChan {
		if this.IsAlive() {
			this.mu.Lock()
			this.conn.Write(this.OutBoundBuffer.Bytes())
			this.OutBoundBuffer.Reset()
			this.mu.Unlock()
		} else {
			return
		}
	}
}

func (this *IoSession) doAsyncTask() {
	for f := range this.AsyncTaskChan {
		if this.IsAlive() {
			f()
		}
	}
}

func (this *IoSession) SetAttribute(k string, v interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Attribute[k] = v

}

func (this *IoSession) GetAttribute(k string) interface{} {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.Attribute[k]
}

func (this *IoSession) Close() {
	if this.IsAlive() {
		this.SetAlive(false)
		this.conn.Close()
	}
}

func (this *IoSession) AddTask(task func()) {
	if this.IsAlive() {
		this.AsyncTaskChan <- task
	}
}
