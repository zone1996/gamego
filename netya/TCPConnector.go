package netya

import (
	"net"

	log "github.com/zone1996/logo"
)

type TCPConnector struct {
	Addr    string
	codec   Codec
	h       Handler
	session *TCPSession
}

func NewTCPConnector(addr string, h Handler, codec Codec) *TCPConnector {
	c := &TCPConnector{
		Addr:  addr,
		codec: codec,
		h:     h,
	}
	return c
}

func (c *TCPConnector) Connect() bool {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return false
	}
	session := NewIoSession(conn)
	c.session = session
	go c.run()
	return true
}

func (c *TCPConnector) run() {
	s := c.session
	data := make([]byte, 1024)
	defer s.Close()
	for {
		n, err := s.conn.Read(data)
		if err != nil {
			log.Error("?", err)
			return
		}
		s.InBoundBuffer.Write(data[:n])
		if pbmsg, err := c.codec.Decode(s.InBoundBuffer); err == nil {
			for _, msg := range pbmsg {
				if msg != nil {
					c.h.OnMessage(s, msg) // do not block here
				}
			}
		} else if err == ErrTooLargeMsg || err == ErrMagicNotRight {
			log.Error("err:?", err)
			return
		} else {
			log.Error("err:?", err)
		}
	}
}

func (c *TCPConnector) Write(b []byte) (n int, err error) {
	return c.session.Write(b)
}

func (c *TCPConnector) AsyncSend(msg *PbMsg) {
	if data, ok := c.codec.Encode(msg); ok {
		c.AsyncWrite(data)
	}
}

func (c *TCPConnector) AsyncWrite(b []byte) {
	c.session.AsyncWrite(b)
}

func (c *TCPConnector) Shutdown() {
	c.session.Close()
}
