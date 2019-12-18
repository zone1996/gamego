package netya

type Handler interface {
	OnConnected(session *TCPSession)
	OnMessage(session *TCPSession, msg *PbMsg)
	OnDisconnected(session *TCPSession)
}

// 对于UDP，没有连接一说:因此 OnConnected、OnDisconnected都未使用
type UdpHandler interface {
	OnConnected(session *UDPSession)
	OnMessage(session *UDPSession, packet []byte)
	OnDisconnected(session *UDPSession)
}
