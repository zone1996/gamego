package netya

import (
	"net"
)

type UDPSession struct {
	id         int64
	conn       *net.UDPConn
	remoteAddr *net.UDPAddr
}

func newUDPSession(id int64, conn *net.Conn, remoteAddr *net.UDPAddr) *UDPSession {
	s := &UDPSession{
		id:         id,
		conn:       conn,
		remoteAddr: remoteAddr,
	}
	return s
}
