package netya

type Handler interface {
	OnConnected(session *IoSession)
	OnMessage(session *IoSession, msg *PbMsg)
	OnDisconnected(session *IoSession)
}
