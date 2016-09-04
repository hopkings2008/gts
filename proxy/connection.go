package proxy

import (
	"net"
	"time"
)

type netConn struct {
	net.Conn
}

func (c *netConn) Close() error {
	return c.Conn.Close()
}

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 0,
}
