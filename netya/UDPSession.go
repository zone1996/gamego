package netya

import (
	"errors"
	"net"
	"sync"
)

type UDPSession struct {
	id         string
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
	closed     bool
	acceptor   *UDPAcceptor
	mu         sync.RWMutex
	attribute  map[string]interface{}
	errClosed  error
}

func newUDPSession(id string, conn *net.UDPConn, remoteAddr *net.UDPAddr, ac *UDPAcceptor) *UDPSession {
	s := &UDPSession{
		id:         id,
		conn:       conn,
		remoteAddr: remoteAddr,
		acceptor:   ac,
		attribute:  make(map[string]interface{}),
		errClosed:  errors.New("Session Closed:" + remoteAddr.String()),
	}
	return s
}

func (us *UDPSession) Write(data []byte) (n int, err error) {
	if us.Closed() {
		return 0, us.errClosed
	}
	return us.conn.WriteToUDP(data, us.remoteAddr)
}

func (us *UDPSession) WriteAsync(b []byte) {
	// TODO implement this
}

func (us *UDPSession) Close() {
	us.mu.RLock()
	if us.closed {
		us.mu.RUnlock()
		return
	}
	us.mu.Lock()
	defer us.mu.Unlock()
	if us.closed {
		return
	}
	us.closed = true
	if us.acceptor != nil {
		us.acceptor.RemoveSession(us.id)
	}
}

func (us *UDPSession) Closed() bool {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.closed
}

func (us *UDPSession) SetAttribute(k string, v interface{}) {
	us.mu.Lock()
	defer us.mu.Unlock()
	us.attribute[k] = v
}

func (us *UDPSession) GetAttribute(k string) interface{} {
	us.mu.RLock()
	defer us.mu.RUnlock()
	return us.attribute[k]
}
