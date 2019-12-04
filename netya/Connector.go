package netya

import (
	"net"

	log "github.com/zone1996/logo"
)

type Connector struct {
	Addr    string
	codec   Codec
	h       Handler
	session *IoSession
}

func NewConnector(addr string, h Handler, codec Codec) *Connector {
	c := &Connector{
		Addr:  addr,
		codec: codec,
		h:     h,
	}
	return c
}

func (c *Connector) Connect() bool {
	conn, err := net.Dial("tcp", c.Addr)
	if err != nil {
		return false
	}
	session := NewIoSession(conn)
	c.session = session
	go c.run()
	return true
}

func (c *Connector) run() {
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
		} else if err == ErrTooLargeMsg {
			log.Error("?", err)
			return
		}
	}
}

func (c *Connector) Write(b []byte) (n int, err error) {
	return c.session.Write(b)
}

func (c *Connector) AsyncSend(msg *PbMsg) {
	c.session.AsyncSend(msg)
}

func (c *Connector) AsyncWrite(b []byte) {
	c.session.AsyncWrite(b)
}

func (c *Connector) Shutdown() {
	c.session.Close()
}
