package netya

type IoHandler interface {
	OnConnected(session IoSession)
	OnMessage(session IoSession, message []byte) // 用于TCP时，程序负责处理半包、粘包
	OnDisconnected(session IoSession)
}

type IoHandlerAdapter struct{}

func (h *IoHandlerAdapter) OnConnected(session IoSession)               {}
func (h *IoHandlerAdapter) OnMessage(session IoSession, message []byte) {}
func (h *IoHandlerAdapter) OnDisconnected(session IoSession)            {}
