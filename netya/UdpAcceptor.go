package netya

import (
	"gamego/tools/gopool"
	"net"

	log "github.com/zone1996/logo"
)

type UdpAcceptor struct {
	codec    UdpCodec
	handler  UdpHandler
	executor gopool.Executor

	config  *AcceptorConfig
	udpAddr *net.UDPAddr
	udpconn *net.UDPConn
}

func NewUdpAcceptor(config *AcceptorConfig, codec UdpCodec, handler UdpHandler, executor gopool.Executor) *UdpAcceptor {
	if config.ReadBufferSize < 1024 {
		config.ReadBufferSize = 1024
	}
	if config.WriteBufferSize < 4096 {
		config.WriteBufferSize = 4096
	}

	uac := &UdpAcceptor{
		codec:    codec,
		handler:  handler,
		executor: executor,
		config:   config,
	}
	return uac
}

func (uac *UdpAcceptor) init() {
	udpAddr, err := net.ResolveUDPAddr("udp", uac.config.Addr)
	if err != nil {
		log.Fatal("?", err)
	}

	udpconn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("?", err)
	}
	udpconn.SetReadBuffer(uac.config.ReadBufferSize)
	udpconn.SetWriteBuffer(uac.config.WriteBufferSize)

	uac.udpAddr = udpAddr
	uac.udpconn = udpconn
}

func (uac *UdpAcceptor) Accept() {
	uac.init()

	data := make([]byte, uac.config.ReadBufferSize)
	for {
		n, remoteAddr, err := uac.udpconn.ReadFromUDP(data)
		if err != nil {
			log.Info("UDPConn Read err:?", err)
			continue
		}
		// TODO 存储UDPSession
		packet := make([]byte, n)
		copy(packet, data[:n])
		f := func() {
			uac.handlePacket(remoteAddr, packet)
		}
		uac.executor.Execute(f)
	}
}

func (uac *UdpAcceptor) handlePacket(remoteAddr *net.UDPAddr, data []byte) {
	pb, err := uac.codec.Decode(data)
	if err != nil {
		log.Info("UdpDecode err:?", err)
		return
	}
	uac.handler.OnMessage(uac.udpconn, remoteAddr, pb)
}

func (uac *UdpAcceptor) Shutdown() {
	uac.executor.Shutdown()
	uac.udpconn.Close()
}
