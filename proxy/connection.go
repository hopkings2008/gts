package proxy

import (
	"net"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type limit struct {
	locker   *sync.Mutex
	maxConn  int32
	maxRps   int32
	conn     int32
	rps      int32
	rpsBegin *time.Time
}

func (l *limit) upConn() bool {
	if l.maxConn == 0 {
		return true
	}

	l.locker.Lock()
	defer l.locker.Unlock()
	l.conn++
	if l.conn <= l.maxConn {
		return true
	}
	return false
}

func (l *limit) downConn() {
	if l.maxConn == 0 {
		return
	}
	l.locker.Lock()
	defer l.locker.Unlock()
	l.conn--
}

func (l *limit) upRps() bool {
	if l.maxRps == 0 {
		return true
	}
	l.locker.Lock()
	defer l.locker.Unlock()
	if l.rpsBegin == nil {
		now := time.Now()
		l.rpsBegin = &now
		l.rps++
		if l.rps <= l.maxRps {
			return true
		}
		return false
	}
	l.rps++
	seconds := time.Since(*l.rpsBegin).Seconds()
	if (int32(float64(l.rps) / seconds)) <= l.maxRps {
		return true
	}
	return false
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
