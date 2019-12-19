package netya

import (
	"net"
	"sync"
	"sync/atomic"

	log "github.com/zone1996/logo"
)

const TCP_SESSION_KEY_IN_BUFFER = "KEY_IN_BYTE_BUFFER"

type TCPSession struct {
	Id                  int64
	conn                net.Conn
	InBoundBuffer       *ByteBuf // for cumulate bytes
	OutBoundBuffer      *ByteBuf // for async write
	AsyncWriteChan      chan struct{}
	AsyncTaskChan       chan func()
	closeAsyncTaskChan  chan struct{}
	closeAsyncWriteChan chan struct{}
	Attribute           map[string]interface{}
	AliveState          int32 // 1:alive
	mu                  sync.RWMutex
}

func NewTCPSession(conn net.Conn) *TCPSession {
	session := &TCPSession{
		conn:                conn,
		InBoundBuffer:       NewByteBuf(1024, 10240),
		OutBoundBuffer:      NewByteBuf(1024, 10240),
		Attribute:           make(map[string]interface{}),
		AsyncWriteChan:      make(chan struct{}),
		AsyncTaskChan:       make(chan func(), 64),
		closeAsyncTaskChan:  make(chan struct{}),
		closeAsyncWriteChan: make(chan struct{}),
		AliveState:          1,
	}
	session.Attribute[TCP_SESSION_KEY_IN_BUFFER] = session.InBoundBuffer
	return session
}

func (this *TCPSession) setId(id int64) {
	this.Id = id
}

func (this *TCPSession) GetId() int64 {
	return this.Id
}

func (this *TCPSession) IsAlive() bool {
	return atomic.LoadInt32(&this.AliveState) == 1
}

func (this *TCPSession) SetAlive(alive bool) {
	var v int32 = 0
	if alive {
		v = 1
	}
	atomic.SwapInt32(&this.AliveState, v)
}

func (this *TCPSession) Write(b []byte) (n int, err error) {
	if this.IsAlive() && b != nil {
		return this.conn.Write(b)
	}
	return 0, nil
}

func (this *TCPSession) WriteAsync(b []byte) {
	if this.IsAlive() && b != nil {
		this.mu.Lock()
		defer this.mu.Unlock()
		n, err := this.OutBoundBuffer.Write(b)
		if err != nil {
			log.Info("WriteAsync err: ?", err)
		}
		if n > 0 {
			this.AsyncWriteChan <- struct{}{}
		}
	}
}

func (this *TCPSession) doAsyncWrite() {
	for {
		select {
		case <-this.AsyncWriteChan:
			if this.IsAlive() {
				this.mu.Lock()
				this.conn.Write(this.OutBoundBuffer.Bytes())
				this.OutBoundBuffer.Reset()
				this.mu.Unlock()
			}
		case <-this.closeAsyncWriteChan:
			break
		}
	}
}

func (this *TCPSession) doAsyncTask() {
	for {
		select {
		case f := <-this.AsyncTaskChan:
			if this.IsAlive() {
				f()
			}
		case <-this.closeAsyncTaskChan:
			break
		}
	}
}

func (this *TCPSession) SetAttribute(k string, v interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.Attribute[k] = v
}

func (this *TCPSession) GetAttribute(k string) interface{} {
	this.mu.RLock()
	defer this.mu.RUnlock()
	return this.Attribute[k]
}

func (this *TCPSession) Close() {
	if this.IsAlive() {
		this.closeAsyncTaskChan <- struct{}{}
		this.closeAsyncWriteChan <- struct{}{}
		this.SetAlive(false)
		this.conn.Close()
	}
}

func (this *TCPSession) Closed() bool {
	return this.IsAlive()
}

func (this *TCPSession) AddTask(task func()) {
	if this.IsAlive() {
		this.AsyncTaskChan <- task
	}
}
