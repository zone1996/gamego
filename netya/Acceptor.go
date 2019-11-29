package netya

import (
	"io"
	"net"
	"time"
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

func (Acceptor) NewAcceptor(config *AcceptorConfig, h Handler, codec Codec) *Acceptor {
	ac := &Acceptor{
		config:  config,
		handler: h,
		codec:   codec,
		conns:   make(map[int32]*IoSession),
	}
	return ac
}

func (ac *Acceptor) init() {

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
		session := IoSession.NewIoSession(conn)
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
	defer h.OnDisconnected(s)

	for {
		data := make([]byte, 1024)
		_, err := s.conn.Read(data)
		if err != nil && err == io.EOF {
			return
		}
		s.InBoundBuffer.Write(data)
		if pbmsg, ok := codec.Decode(s.InBoundBuffer.Bytes()); ok {
			s.InBoundBuffer.Truncate(0)
			for msg := range pbmsg {
				h.OnMessage(s, msg)
			}
		}
	}
}

func (ac *Acceptor) Shutdown() {

}
