package netya

import (
	"net"
	"time"

	log "github.com/zone1996/logo"
)

type AcceptorConfig struct {
	Addr            string // ":6666"
	ReadBufferSize  int
	WriteBufferSize int
	WSPath          string // eg: "/echo", Addr:"localhost:8080"
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

func (ac *Acceptor) init() error {
	listener, err := net.Listen("tcp", ac.config.Addr)
	if err != nil {
		return err
	}
	ac.listener = listener
	return nil
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

	go s.doAsyncWrite()
	go s.doAsyncTask()

	data := make([]byte, 1024)
	for {
		n, err := s.conn.Read(data)
		if err != nil {
			log.Error("Err:?", err)
			return
		}
		_, err = s.InBoundBuffer.Write(data[:n])
		if err != nil {
			log.Error("Err:?", err)
			return
		}
		if pbmsg, err := codec.Decode(s.InBoundBuffer); err == nil {
			for _, msg := range pbmsg {
				if msg != nil {
					h.OnMessage(s, msg) // do not block here
				}
			}
		} else if err == ErrTooLargeMsg || err == ErrMagicNotRight {
			log.Error("Err:?", err)
			return
		} else {
			log.Info("Err:?", err)
		}
	}
}

func (ac *Acceptor) Shutdown() {
	ac.listener.Close()
	for _, s := range ac.conns {
		s.Close()
	}
}
