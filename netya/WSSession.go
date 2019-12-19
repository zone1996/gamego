package netya

import (
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var ErrWSConnClosed = errors.New("WSConn closed")

type WSSession struct {
	id        int64
	conn      *websocket.Conn
	attribute map[string]interface{} // 可以存储玩家ID等信息
	closed    bool
	mu        sync.RWMutex
}

func newWSSession(id int64, conn *websocket.Conn) *WSSession {
	s := &WSSession{
		id:        id,
		conn:      conn,
		attribute: make(map[string]interface{}),
	}
	return s
}

func (this *WSSession) Write(msg []byte) (n int, err error) {
	if this.closed {
		return 0, ErrWSConnClosed
	}

	this.mu.Lock()
	defer this.mu.Unlock()
	if !this.closed {
		err := this.conn.WriteMessage(websocket.BinaryMessage, msg)
		if err == nil {
			return len(msg), nil
		}
		return 0, err
	}
	return 0, ErrWSConnClosed
}

func (us *WSSession) WriteAsync(b []byte) {
	// TODO implement this
}

// 代码中主动关闭:希望主动移除某个session时，如将玩家踢下线
// 如果在服务端Close，ServeHTTP中的for循环也将退出
func (this *WSSession) Close() {
	this.mu.RLock()
	if this.closed {
		this.mu.RUnlock()
		return
	}

	this.mu.Lock()
	defer this.mu.Unlock()
	if !this.closed {
		this.conn.WriteControl(websocket.CloseMessage, []byte("close"), time.Now().Add(time.Second))
		this.closed = true
		this.conn.Close()
	}
}

func (this *WSSession) Closed() bool {
	return this.closed
}

// 不要并发调用
func (this *WSSession) readMessage() (messageType int, p []byte, err error) {
	if this.closed {
		return -1, nil, ErrWSConnClosed
	}
	return this.conn.ReadMessage()
}

func (this *WSSession) SetAttribute(k string, v interface{}) {
	this.mu.Lock()
	defer this.mu.Unlock()
	this.attribute[k] = v
}

func (this *WSSession) GetAttribute(k string) interface{} {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.attribute[k]
}
