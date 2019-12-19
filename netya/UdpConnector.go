package netya

import (
	"net"
	"sync"

	log "github.com/zone1996/logo"
)

type UdpConnector struct {
	handler IoHandler
	config  *AcceptorConfig
	session IoSession
	closed  bool
	mu      sync.RWMutex
}

func NewUdpConnector(config *AcceptorConfig, h IoHandler) *UdpConnector {
	uc := &UdpConnector{
		config:  config,
		handler: h,
	}
	return uc
}

func (uc *UdpConnector) Connect() bool {
	serverAddr, err := net.ResolveUDPAddr("udp", uc.config.Addr)
	if err != nil {
		log.Error("UDP Connect err:?", err)
		return false
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Error("UDP Connect err:?", err)
		return false
	}

	conn.SetReadBuffer(uc.config.ReadBufferSize)
	conn.SetWriteBuffer(uc.config.WriteBufferSize)
	uc.session = newUDPSession("", conn, serverAddr, nil)
	uc.handler.OnConnected(uc.session)
	go uc.run()
	return true
}

func (uc *UdpConnector) run() {
	session := uc.session.(*UDPSession)
	data := make([]byte, uc.config.ReadBufferSize)
	for !uc.Closed() {
		n, _, err := session.conn.ReadFromUDP(data)
		if err != nil {
			continue
		}
		packet := make([]byte, n)
		copy(packet, data[:n])
		f := func() {
			uc.handler.OnMessage(session, packet)
		}
		f() // or using executor:尽量减少因程序原因造成丢包
	}
}

func (uc *UdpConnector) WriteBytes(data []byte) {
	if uc.Closed() {
		return
	}
	if len(data) > uc.config.WriteBufferSize {
		// log
		return
	}
	uc.session.Write(data)
}

func (uc *UdpConnector) Shutdown() {
	if uc.closed {
		return
	}

	uc.mu.Lock()
	defer uc.mu.Lock()
	uc.session.Close()
	uc.closed = true
	uc.handler.OnDisconnected(uc.session)
}

func (uc *UdpConnector) Closed() bool {
	uc.mu.RLock()
	defer uc.mu.RUnlock()
	return uc.closed
}
