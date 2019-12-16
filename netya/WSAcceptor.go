package netya

import (
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type WSAcceptor struct {
	config   *AcceptorConfig
	server   *http.Server
	upgrader websocket.Upgrader
	conns    map[int64]*websocket.Conn

	mu     sync.Mutex
	closed bool
	idGen  int64
}

func NewWSAcceptor(config *AcceptorConfig) {
	wsa := &WSAcceptor{
		config: config,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
		},
		conns: make(map[int64]*websocket.Conn),
	}
}

type wshandler struct {
	wsa *WSAcceptor
}

func (wsh *wshandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	wsa := wsh.wsa
	conn, err := wsa.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	// TODO conn config

	id := wsa.addConn(conn)
	defer wsa.removeConn(id)

	for !wsa.closed {
		msgType, b, err := conn.ReadMessage()
		if err != nil || msgType != websocket.BinaryMessage {
			return
		}
		// TODO decode
		// execute
		conn.WriteMessage(msgType, b)
	}
}

func (wsa *WSAcceptor) addConn(conn *websocket.Conn) (id int64) {
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	id = wsa.idGen
	wsa.conns[wsa.idGen] = conn
	wsa.idGen++
	return
}

func (wsa *WSAcceptor) removeConn(id int64) {
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	delete(wsa.conns, id)
}

func (wsa *WSAcceptor) Accept() {
	server := &http.Server{
		Addr:    wsa.config.Addr,
		Handler: &wshandler{wsa: wsa},
	}
	server.ListenAndServe()
	wsa.server = server
}

func (wsa *WSAcceptor) Shutdown() {
	wsa.mu.Lock()
	defer wsa.mu.Unlock()
	if wsa.closed {
		return
	}
	wsa.server.Close()
	wsa.closed = true
}
