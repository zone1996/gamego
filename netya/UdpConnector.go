package netya

import (
	"net"

	log "github.com/zone1996/logo"
)

type UdpConnector struct {
	codec      UdpCodec
	handler    UdpHandler
	ServerAddr string
	RemoteAddr *net.UDPAddr
	conn       *net.UDPConn
	closed     bool
}

func NewUdpConnector(serverAddr string, h UdpHandler, codec UdpCodec) *UdpConnector {
	uc := &UdpConnector{
		ServerAddr: serverAddr,
		handler:    h,
		codec:      codec,
	}
	return uc
}

func (uc *UdpConnector) Connect() bool {
	serverAddr, err := net.ResolveUDPAddr("udp", uc.ServerAddr)
	if err != nil {
		log.Error("UDP Connect err:?", err)
		return false
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Error("UDP Connect err:?", err)
		return false
	}

	conn.SetReadBuffer(4096)
	conn.SetWriteBuffer(4096)
	uc.conn = conn
	uc.RemoteAddr = serverAddr
	go uc.run()
	return true
}

func (uc *UdpConnector) run() {
	data := make([]byte, 4096)
	for !uc.closed {
		n, remoteAddr, err := uc.conn.ReadFromUDP(data)
		if err != nil {
			continue
		}
		pb, err := uc.codec.Decode(data[:n])
		if err != nil {
			continue
		}
		f := func() {
			uc.handler.OnMessage(uc.conn, remoteAddr, pb)
		}
		f() // or using executor:尽量减少因程序原因造成丢包
	}
}

func (uc *UdpConnector) WriteBytes(data []byte, remoteAddr *net.UDPAddr) {
	uc.conn.WriteToUDP(data, remoteAddr)
}

func (uc *UdpConnector) WritePbMsg(pb *PbMsg, remoteAddr *net.UDPAddr) {
	if data, ok := uc.codec.Encode(pb); ok {
		uc.conn.WriteToUDP(data, remoteAddr)
	}
}

func (uc *UdpConnector) Shutdown() {
	uc.conn.Close()
	uc.closed = true
}
