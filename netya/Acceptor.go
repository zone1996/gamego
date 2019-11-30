package netya

import (
	"io"
	"net"
	"time"

	log "github.com/zone1996/logo"
)

type AcceptorConfig struct {
	Port string
}

type Acceptor struct {
	config   *AcceptorConfig
	listener net.Listener
	handler  Handler
	codec    Codec

	sleepDuration      time.Duration
	sessionIdGenerator int32
	conns              map[int32]*IoSession
}

func NewAcceptor(config *AcceptorConfig, h Handler, codec Codec) *Acceptor {
	ac := &Acceptor{
		config:             config,
		handler:            h,
		codec:              codec,
		conns:              make(map[int32]*IoSession),
		sleepDuration:      time.Second,
		sessionIdGenerator: 0,
	}
	return ac
}

func (ac *Acceptor) init() {
	listener, err := net.Listen("tcp", ac.config.Port)
	if err != nil {
		log.Fatal("?", err)
	}
	ac.listener = listener
}

func (ac *Acceptor) temporarySleep() {
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

func (ac *Acceptor) Accept() {
	ac.init()
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
		session := NewIoSession(conn)
		session.SetId(ac.sessionIdGenerator)
		go runSession(session, ac)

		ac.sessionIdGenerator += 1
		ac.conns[session.Id] = session
	}
}

func runSession(s *IoSession, ac *Acceptor) {
	codec := ac.codec
	h := ac.handler
	h.OnConnected(s)
	defer func() {
		h.OnDisconnected(s)
		delete(ac.conns, s.Id) // maybe concurrenttly mod conns
	}()
	go func() {
		for {
			if s.OutBoundBuffer.Len() > 0 {
				if s.IsAlive() {
					s.mu.Lock()
					s.conn.Write(s.OutBoundBuffer.Bytes())
					s.OutBoundBuffer.Reset()
					s.mu.Unlock()
				} else {
					return
				}
			}
		}
	}()

	for {
		data := make([]byte, 2048)
		_, err := s.conn.Read(data)
		if err != nil && err == io.EOF {
			return
		}
		s.InBoundBuffer.Write(data)
		if pbmsg, ok := codec.Decode(s.InBoundBuffer); ok {
			for _, msg := range pbmsg {
				h.OnMessage(s, msg) // do not block here
			}
		}
	}
}

func (ac *Acceptor) Shutdown() {
	ac.listener.Close()
	for _, s := range ac.conns {
		s.Close()
	}
}
