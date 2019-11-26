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
	// 编解码器

	sleepDuration      time.Duration
	sessionIdGenerator int32
	conns              map[int32]*IoSession
}

func (Acceptor) NewAcceptor(config *AcceptorConfig, h Handler) *Acceptor {
	ac := &Acceptor{
		config:  config,
		handler: h,
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
		go runSession(session, ac.handler)

		ac.sessionIdGenerator += 1
		ac.conns[session.Id] = session
	}
}

func runSession(s *IoSession, h Handler) {
	h.OnConnected(s)
	defer h.OnDisconnected(s)

	for {
		header := make([]byte, HEADER_SIZE)
		_, err := io.ReadFull(s.conn, header)
		if err != nil {
			s.Close()
			return
		}
		// TODO 解码
		var code int16 = 1
		pbmsg := PbMsg.NewPbMsg(code)
		h.OnMessage(s, pbmsg)
	}
}

func (ac *Acceptor) Shutdown() {

}
