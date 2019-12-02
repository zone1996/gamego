package main

import (
	cmds "gamego/cmd"
	"gamego/netya"

	log "github.com/zone1996/logo"
)

type DefaultHandler struct {
	acceptor *netya.Acceptor
}

func (h *DefaultHandler) OnConnected(session *netya.IoSession) {
	log.Info("Session ? Connected.", session.Id)
}

func (h *DefaultHandler) OnMessage(session *netya.IoSession, msg *netya.PbMsg) {
	code := msg.Code

	if c, ok := cmds.GetCmd(code); ok {
		session.AddTask(wrapTask(c, session, msg))
	} else {
		log.Info("Cmd not found, code:?", code)
	}
}

func (h *DefaultHandler) OnDisconnected(session *netya.IoSession) {
	session.Close()
	log.Info("Session ? DisConnected.", session.Id)
}

func wrapTask(cmd cmds.Cmd, session *netya.IoSession, msg *netya.PbMsg) func() {
	f := func() {
		cmd.Exec(session, msg)
	}
	return f
}
