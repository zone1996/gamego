package netya

type Acceptor interface {
	Network() string // tcp|udp|ws
	Accept()         // block method
	RemoveSession(key interface{}) error
	Shutdown()
}

type Connector interface {
	Connect() bool // non-blocking
	Write(b []byte) (n int, err error)
	WriteAsync(b []byte)
	Shutdown()
}
