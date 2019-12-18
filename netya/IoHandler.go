package netya

type IoHandler interface {
	OnConnected(session IoSession)
	OnMessage(session IoSession, message []byte) // TCP协议解包时，拿到完整包再调用
	OnDisconnected(session IoSession)
}
