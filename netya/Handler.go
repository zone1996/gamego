package netya

import (
	"net"
)

type Handler interface {
	OnConnected(session *IoSession)
	OnMessage(session *IoSession, msg *PbMsg)
	OnDisconnected(session *IoSession)
}

type UdpHandler interface {
	// handle msg, you can using udpconn to write msg back to remoteAddr
	OnMessage(udpconn *net.UDPConn, remoteAddr *net.UDPAddr, msg *PbMsg)
}
