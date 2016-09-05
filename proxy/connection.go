package proxy

import (
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type limit struct {
	locker  *sync.Mutex
	maxConn int32
	maxRps  int32
	conn    int32
	rps     int32
}

func (l *limit) upConn() bool {
	l.locker.Lock()
	defer l.locker.Unlock()
	if l.maxConn == 0 {
		return true
	}
	if l.conn <= l.maxConn {
		l.conn++
		return true
	}
	return false
}

func (l *limit) downConn() {
	l.locker.Lock()
	defer l.locker.Unlock()
	if l.maxConn != 0 {
		l.conn--
	}
}

type netConn struct {
	net.Conn
	limit *limit
}

func (c *netConn) Close() error {
	c.limit.downConn()
	log.Infof("Close %s %s", c.Conn.RemoteAddr().Network(), c.Conn.RemoteAddr().String())
	return c.Conn.Close()
}

func newNetConn(c net.Conn, l *limit) *netConn {
	nc := &netConn{}
	nc.Conn = c
	nc.limit = l
	return nc
}

var dialer = &net.Dialer{
	Timeout:   30 * time.Second,
	KeepAlive: 0,
}
