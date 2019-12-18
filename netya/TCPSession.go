package netya

import (
	"net"
	"sync"
	"sync/atomic"

	log "github.com/zone1996/logo"
)

type TCPSession struct {
	Id                  int32
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
		InBoundBuffer:       NewByteBuf(1024, 2*int(MAX_PACKET_SIZE)),
		OutBoundBuffer:      NewByteBuf(1024, 2*int(MAX_PACKET_SIZE)),
		Attribute:           make(map[string]interface{}),
		AsyncWriteChan:      make(chan struct{}),
		AsyncTaskChan:       make(chan func(), 64),
		closeAsyncTaskChan:  make(chan struct{}),
		closeAsyncWriteChan: make(chan struct{}),
		AliveState:          1,
	}
	return session
}

func (this *TCPSession) SetId(id int32) {
	this.Id = id
}

func (this *TCPSession) GetId() int32 {
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

func (this *TCPSession) AsyncSend(msg *PbMsg) {
	this.AsyncWrite(msg.Bytes())
}

func (this *TCPSession) AsyncWrite(b []byte) {
	if this.IsAlive() && b != nil {
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

func (this *TCPSession) AddTask(task func()) {
	if this.IsAlive() {
		this.AsyncTaskChan <- task
	}
}
