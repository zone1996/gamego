package example

import (
	"fmt"
	"os"
	"os/signal"
	_ "time"

	"gamego/netya"
	"gamego/pb"
)

type defaultHandler struct {
}

func (h *defaultHandler) OnConnected(session netya.IoSession) {
	fmt.Println("Session ? Connected.", session.(*netya.TCPSession).Id)
}

func (h *defaultHandler) OnMessage(session netya.IoSession, data []byte) {
	fmt.Println("Receive Server data.code=")
}

func (h *defaultHandler) OnDisconnected(session netya.IoSession) {
	fmt.Println("Session ? DisConnected.", session.(*netya.TCPSession).Id)
}

func runExample() {
	c := netya.NewTCPConnector("localhost:6666", &defaultHandler{})
	if !c.Connect() {
		return
	}

	msg := pb.NewPbMsg(1)
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
