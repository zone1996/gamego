package example

import (
	"fmt"
	"os"
	"os/signal"
	_ "time"

	"gamego/netya"
)

type defaultHandler struct{}

func (h *defaultHandler) OnConnected(session *netya.IoSession) {
	fmt.Println("Session ? Connected.", session.Id)
}

func (h *defaultHandler) OnMessage(session *netya.IoSession, msg *netya.PbMsg) {
	fmt.Println("Receive Server data.code=", msg.GetCode(), "userId=", msg.GetUserId(), "payload=?", string(msg.GetPayload()))
}

func (h *defaultHandler) OnDisconnected(session *netya.IoSession) {
	session.Close()
	fmt.Println("Session ? DisConnected.", session.Id)
}

func runExample() {
	c := netya.NewConnector("localhost:6666", &defaultHandler{}, &netya.DefaultCodec{})
	if !c.Connect() {
		return
	}

	msg := netya.NewPbMsg(1)
	msg.SetUserId(666)
	msg.SetValue(0, 100)
	msg.SetPayload([]byte("Payload Data"))
	d := msg.Bytes()
	fmt.Println("d len = ", len(d))
	if d != nil {
		for i := 0; i < 10; i++ {
			n, err := c.Write(d)
			fmt.Println("n=", n, ",err=", err)
		}
	}
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, os.Kill)
	<-stop
}
