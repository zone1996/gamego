package main

import (
	cmds "gamego/cmd"
	"gamego/netya"
	"gamego/pb"

	log "github.com/zone1996/logo"
)

type DefaultHandler struct {
	netya.IoHandler
	codec    DefaultCodec
	acceptor netya.Acceptor
}

func (h *DefaultHandler) OnConnected(session netya.IoSession) {
	s, ok := session.(*netya.TCPSession)
	if !ok {
		session.Close()
		return
	}
	log.Info("Session ? Connected.", s.Id)
}

func (h *DefaultHandler) OnMessage(session netya.IoSession, message []byte) {
	inBoundBuf := session.(*netya.TCPSession).InBoundBuffer
	if _, err := inBoundBuf.Write(message); err != nil {
		session.Close()
		return
	}
	if msgs, err := h.codec.Decode(inBoundBuf); msgs != nil && err == nil {
		for _, msg := range msgs {
			code := msg.GetCode()
			if c, ok := cmds.GetCmd(code); ok {
				session.(*netya.TCPSession).AddTask(wrapTask(c, session, msg))
			} else {
				log.Info("Cmd not found, code:?", code)
			}
		}
	} else if err != nil {
		log.Info("Err:?,", err)
		if err == ErrMagicNotRight || err == ErrTooLargeMsg {
			session.Close()
			return
		}
	}
}

func (h *DefaultHandler) OnDisconnected(session netya.IoSession) {
	log.Info("Session ? DisConnected.", session.(*netya.TCPSession).Id)
}

func wrapTask(cmd cmds.Cmd, session netya.IoSession, msg *pb.PbMsg) func() {
	f := func() {
		cmd.Exec(session, msg)
	}
	return f
}
