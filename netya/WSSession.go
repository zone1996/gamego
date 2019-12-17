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
	Attribute map[string]interface{} // 可以存储玩家ID等信息
	closed    bool
	mu        sync.Mutex
}

func newWSSession(id int64, conn *websocket.Conn) *WSSession {
	s := &WSSession{
		id:        id,
		conn:      conn,
		Attribute: make(map[string]interface{}),
	}
	return s
}

func (this *WSSession) Write(msg []byte) error {
	if this.closed {
		return ErrWSConnClosed
	}

	this.mu.Lock()
	defer this.mu.Unlock()
	if !this.closed {
		return this.conn.WriteMessage(websocket.BinaryMessage, msg)
	}
	return nil
}

func (this *WSSession) Writes(msgs ...[]byte) error {
	if this.closed {
		return ErrWSConnClosed
	}

	this.mu.Lock()
	defer this.mu.Unlock()
	if this.closed {
		return ErrWSConnClosed
	}

	var err error
	for _, msg := range msgs {
		e := this.conn.WriteMessage(websocket.BinaryMessage, msg)
		if e != nil {
			err = e
		}
	}
	return err
}

// 代码中主动关闭:希望主动移除某个session时，如将玩家踢下线
// 如果在服务端Close，ServeHTTP中的for循环也将退出
func (this *WSSession) Close() {
	if this.closed {
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
	this.Attribute[k] = v
}

func (this *WSSession) GetAttribute(k string) interface{} {
	this.mu.Lock()
	defer this.mu.Unlock()
	return this.Attribute[k]
}
