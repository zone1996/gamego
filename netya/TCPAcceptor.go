package netya

import (
	"net"
	"sync"
	"time"

	log "github.com/zone1996/logo"
)

type AcceptorConfig struct {
	Addr            string // ":6666"
	ReadBufferSize  int
	WriteBufferSize int
	WSPath          string // eg: "/echo", Addr:"localhost:8080"
}

type TCPAcceptor struct {
	config   *AcceptorConfig
	listener net.Listener
	handler  IoHandler

	sleepDuration      time.Duration
	sessionIdGenerator int64
	conns              map[int64]*TCPSession
	mu                 sync.RWMutex
}

func NewTCPAcceptor(config *AcceptorConfig, h IoHandler) Acceptor {
	ac := &TCPAcceptor{
		config:             config,
		handler:            h,
		conns:              make(map[int64]*TCPSession),
		sleepDuration:      time.Second,
		sessionIdGenerator: 0,
	}
	return ac
}
func (ac *TCPAcceptor) Network() string {
	return "tcp"
}

func (ac *TCPAcceptor) init() error {
	listener, err := net.Listen("tcp", ac.config.Addr)
	if err != nil {
		return err
	}
	ac.listener = listener
	return nil
}

func (ac *TCPAcceptor) temporarySleep() {
	if ac.sleepDuration == 0 {
		ac.sleepDuration = 5 * time.Millisecond
	} else {
		ac.sleepDuration *= 2
	}
	if ac.sleepDuration >= time.Second {
		ac.sleepDuration = time.Second
	}
	time.Sleep(ac.sleepDuration)
}

func (ac *TCPAcceptor) Accept() {
	if err := ac.init(); err != nil {
		log.Error("Acceptor init err:?", err)
		return
	}
	for {
		conn, err := ac.listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				ac.temporarySleep()
				continue
			}
			return
		}
		ac.sleepDuration = 0
		session := NewTCPSession(conn)
		session.setId(ac.sessionIdGenerator)
		go ac.runSession(session)

		ac.sessionIdGenerator += 1
		ac.conns[session.Id] = session
	}
}

func (ac *TCPAcceptor) runSession(s *TCPSession) {
	h := ac.handler
	h.OnConnected(s)
	defer func() {
		h.OnDisconnected(s)
		ac.RemoveSession(s.Id)
	}()

	go s.doAsyncWrite()
	go s.doAsyncTask()

	data := make([]byte, 1024)
	for {
		n, err := s.conn.Read(data)
		if err != nil {
			log.Error("Err:?", err)
			return
		}
		h.OnMessage(s, data[:n])
	}
}

func (ac *TCPAcceptor) RemoveSession(key interface{}) error {
	k, ok := key.(int64)
	if !ok {
		return nil
	}
	ac.mu.RLock()
	if _, ok := ac.conns[k]; !ok {
		ac.mu.RUnlock()
		return nil
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	delete(ac.conns, k)
	return nil
}

func (ac *TCPAcceptor) Shutdown() {
	ac.listener.Close()
	for _, s := range ac.conns {
		s.Close()
	}
}
