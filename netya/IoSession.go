package netya

import (
	"net"
	"sync"
	"sync/atomic"

	"github.com/gogo/protobuf/proto"

	log "github.com/zone1996/logo"
)

type IoSession struct {
	Id                  int32
	conn                net.Conn
	InBoundBuffer       *ByteBuf // using for cumulate bytes
	OutBoundBuffer      *ByteBuf // using for async write
	AsyncWriteChan      chan struct{}
	AsyncTaskChan       chan func()
	closeAsyncTaskChan  chan struct{}
	closeAsyncWriteChan chan struct{}
	Attribute           map[string]interface{}
	AliveState          int32 // 1:alive
	mu                  sync.RWMutex
}

func NewIoSession(conn net.Conn) *IoSession {
	session := &IoSession{
		conn:                conn,
		InBoundBuffer:       NewByteBuf(1024, 2*MAX_PACKET_SIZE),
		OutBoundBuffer:      NewByteBuf(1024, 2*MAX_PACKET_SIZE),
		Attribute:           make(map[string]interface{}),
		AsyncWriteChan:      make(chan struct{}),
		AsyncTaskChan:       make(chan func(), 64),
		closeAsyncTaskChan:  make(chan struct{}),
		closeAsyncWriteChan: make(chan struct{}),
		AliveState:          1,
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

func (this *IoSession) AsyncSend(msg *PbMsg) {
	msg.Length = int32(msg.XXX_Size())
	data, err := proto.Marshal(msg)
	if err != nil {
		log.Info("proto marshal err:?", err)
		return
	}
	this.AsyncWrite(data)
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

func (this *IoSession) doAsyncTask() {
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
		this.closeAsyncTaskChan <- struct{}{}
		this.closeAsyncWriteChan <- struct{}{}
		this.SetAlive(false)
		this.conn.Close()
	}
}

func (this *IoSession) AddTask(task func()) {
	if this.IsAlive() {
		this.AsyncTaskChan <- task
	}
}
