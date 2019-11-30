package main

import (
	cs "gamego/cmd"
	"gamego/netya"

	log "github.com/zone1996/logo"
)

type DefaultHandler struct{}

func (h *DefaultHandler) OnConnected(session *netya.IoSession) {
	log.Info("Session ? Connected.", session.Id)
}

func (h *DefaultHandler) OnMessage(session *netya.IoSession, msg *netya.PbMsg) {
	code := msg.Code

	if c, ok := cs.GetCmd(code); ok {
		c.Exec(session, msg)
	} else {
		log.Info("Cmd not found, code:?", code)
	}
}

func (h *DefaultHandler) OnDisconnected(session *netya.IoSession) {
	session.Close()
	log.Info("Session ? DisConnected.", session.Id)
}
