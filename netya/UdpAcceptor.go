package netya

import (
	"gamego/tools/gopool"
	"net"
	"sync"
	"syscall"

	log "github.com/zone1996/logo"
)

type UdpAcceptor struct {
	codec    UdpCodec
	handler  UdpHandler
	executor gopool.Executor

	config  *AcceptorConfig
	udpAddr *net.UDPAddr
	udpconn *net.UDPConn

	mu       sync.RWMutex
	sessions map[string]*UDPSession
}

func NewUdpAcceptor(cf *AcceptorConfig, cc UdpCodec, h UdpHandler, e gopool.Executor) *UdpAcceptor {
	uac := &UdpAcceptor{
		codec:    cc,
		handler:  h,
		executor: e,
		config:   cf,
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
			if err == syscall.EINVAL {
				return
			}
			continue
		}

		packet := make([]byte, n)
		copy(packet, data[:n])
		f := func() { // 同一客户端的包可能会并发执行，可以在业务逻辑中顺序执行(ActionQueue)
			session := uac.tryGetSession(remoteAddr)
			uac.handler.OnMessage(session, packet)
		}
		uac.executor.Execute(f)
	}
}

func (uac *UdpAcceptor) tryGetSession(remoteAddr *net.UDPAddr) *UDPSession {
	remoteAddrStr := remoteAddr.String()
	uac.mu.RLock()
	if s, ok := uac.sessions[remoteAddrStr]; ok {
		uac.mu.RUnlock()
		return s
	}
	uac.mu.Lock()
	defer uac.mu.Unlock()
	if s, ok := uac.sessions[remoteAddrStr]; ok {
		return s
	} else {
		s := newUDPSession(remoteAddrStr, uac.udpconn, remoteAddr, uac)
		uac.sessions[remoteAddrStr] = s
		return s
	}
}

func (uac *UdpAcceptor) RemoveSession(key interface{}) error {
	k, ok := key.(string)
	if !ok {
		return nil
	}
	uac.mu.RLock()
	if _, ok := uac.sessions[k]; !ok {
		uac.mu.RUnlock()
		return nil
	}
	uac.mu.Lock()
	defer uac.mu.Unlock()
	delete(uac.sessions, k)
	return nil
}

func (uac *UdpAcceptor) Shutdown() {
	uac.executor.Shutdown()
	uac.udpconn.Close()
}
