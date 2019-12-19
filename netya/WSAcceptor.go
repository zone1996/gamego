package netya

import (
	"gamego/tools/gopool"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type WSHandler interface {
	OnConnected(session *WSSession)
	OnMessage(session *WSSession, msgType int, data []byte)
	OnDisconnected(session *WSSession)
}

type WSAcceptor struct {
	config   *AcceptorConfig
	server   *http.Server
	upgrader websocket.Upgrader
	sessions map[int64]*WSSession
	handler  WSHandler
	executor gopool.Executor

	mu     sync.Mutex
	closed bool
	idGen  int64
}

func NewWSAcceptor(config *AcceptorConfig, h WSHandler, e gopool.Executor) Acceptor {
	wsa := &WSAcceptor{
		config: config,
		upgrader: websocket.Upgrader{
			HandshakeTimeout: 3 * time.Second,
			ReadBufferSize:   config.ReadBufferSize,
			WriteBufferSize:  config.WriteBufferSize,
		},
		sessions: make(map[int64]*WSSession),
		handler:  h,
		executor: e,
	}
	return wsa
}

type wsUpgradeHandler struct {
	wsa *WSAcceptor
}

func (wuh *wsUpgradeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsa := wuh.wsa
	conn, err := wsa.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// TODO conn config

	id := wsa.genId()
	session := newWSSession(id, conn)
	wsa.addSession(session)
	wsa.handler.OnConnected(session)
	defer func() {
		session.Close()
		wsa.handler.OnDisconnected(session)
		wsa.RemoveSession(id)
	}()

	for !wsa.closed {
		msgType, msg, err := session.readMessage() // one goroutine read message
		if err != nil {                            // maybe client closed
			return
		}
		f := func() {
			wsa.handler.OnMessage(session, msgType, msg)
		}
		wsa.executor.Execute(f)
	}
}

func (wsa *WSAcceptor) addSession(session *WSSession) {
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	wsa.sessions[session.id] = session
	return
}

func (wsa *WSAcceptor) genId() int64 {
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	wsa.idGen += 1
	return wsa.idGen
}

func (wsa *WSAcceptor) Network() string {
	return "ws"
}

func (wsa *WSAcceptor) RemoveSession(key interface{}) error {
	id := key.(int64)
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	delete(wsa.sessions, id)
	return nil
}

func (wsa *WSAcceptor) Accept() {
	server := &http.Server{
		Addr:    wsa.config.Addr,
		Handler: &wsUpgradeHandler{wsa: wsa},
	}
	// 证书、balabala
	wsa.server = server
	server.ListenAndServe()
}

func (wsa *WSAcceptor) Shutdown() {
	if wsa.closed {
		return
	}
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	if wsa.closed {
		return
	}
	wsa.server.Close()
	wsa.closed = true
}
