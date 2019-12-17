package netya

import (
	"fmt"
	"gamego/tools/gopool"
	"net/url"

	"github.com/gorilla/websocket"
)

type WSConnector struct {
	config   *AcceptorConfig
	session  *WSSession
	handler  WSHandler
	executor gopool.Executor
}

func NewWSConnector(config *AcceptorConfig, h WSHandler, e gopool.Executor) *WSConnector {
	c := &WSConnector{
		config:   config,
		handler:  h,
		executor: e,
	}
	return c
}

func (wsc *WSConnector) Connect() bool {
	u := url.URL{
		Scheme: "ws",
		Host:   wsc.config.Addr,
		Path:   wsc.config.WSPath,
	}
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("WS Connect err:", err)
		return false
	}
	go wsc.run(conn)
	return true
}

func (wsc *WSConnector) run(conn *websocket.Conn) {
	sess := newWSSession(1, conn)
	wsc.handler.OnConnected(sess)
	defer wsc.handler.OnDisconnected(sess)
	for {
		msgType, msg, err := wsc.session.readMessage()
		if err != nil {
			wsc.Close()
			return
		}
		f := func() {
			wsc.handler.OnMessage(sess, msgType, msg)
		}
		if wsc.executor != nil {
			wsc.executor.Execute(f)
		} else {
			f()
		}
	}
}

func (wsc *WSConnector) Writes(msgs ...[]byte) {
	if err := wsc.session.Writes(msgs...); err != nil {
		fmt.Println("WS Writes err:", err)
	}
}

func (wsc *WSConnector) Write(msg []byte) {
	if err := wsc.session.Write(msg); err != nil {
		fmt.Println("WS Write err:", err)
	}
}

func (wsc *WSConnector) Close() {
	wsc.session.Close()
}

func (wsc *WSConnector) Closed() bool {
	return wsc.session.Closed()
}
