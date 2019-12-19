package netya

import (
	"net"

	log "github.com/zone1996/logo"
)

type TCPConnector struct {
	Addr    string
	h       IoHandler
	session *TCPSession
}

func NewTCPConnector(addr string, h IoHandler) *TCPConnector {
	c := &TCPConnector{
		Addr: addr,
		h:    h,
	}
	return c
}

func (c *TCPConnector) Connect() bool {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return false
	}
	session := NewTCPSession(conn)
	c.session = session
	go c.run()
	return true
}

func (c *TCPConnector) run() {
	s := c.session
	data := make([]byte, 1024)
	defer func() {
		s.Close()
		c.h.OnDisconnected(s)
	}()
	for {
		n, err := s.conn.Read(data)
		if err != nil {
			log.Error("?", err)
			return
		}
		c.h.OnMessage(s, data[:n])
	}
}

func (c *TCPConnector) Write(b []byte) (n int, err error) {
	return c.session.Write(b)
}

func (c *TCPConnector) AsyncWrite(b []byte) {
	c.session.WriteAsync(b)
}

func (c *TCPConnector) Shutdown() {
	c.session.Close()
}
