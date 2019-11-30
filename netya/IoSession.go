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
	InBoundBuffer  *bytes.Buffer // using for cumulative bytes
	OutBoundBuffer *bytes.Buffer // using for async write
	Attribute      map[string]interface{}
	AliveState     int32 // 1:存活
	mu             sync.RWMutex
}

func NewIoSession(conn net.Conn) *IoSession {
	session := &IoSession{
		conn:           conn,
		InBoundBuffer:  new(bytes.Buffer),
		OutBoundBuffer: new(bytes.Buffer),
		Attribute:      make(map[string]interface{}),
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
		_, err := this.OutBoundBuffer.Write(b)
		if err != nil {
			log.Info("AsyncWrite err: ?", err)
		}
	}
}

func (this *IoSession) SetAttribute(k string, v interface{}) {
	this.Attribute[k] = v
}

func (this *IoSession) GetAttribute(k string) interface{} {
	return this.Attribute[k]
}

func (this *IoSession) Close() {
	if this.IsAlive() {
		this.SetAlive(false)
		this.conn.Close()
	}
}
