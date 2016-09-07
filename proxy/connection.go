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

func (l *limit) upConn() {
	if l.maxConn == 0 {
		return
	}

	l.locker.Lock()
	l.conn++
	l.locker.Unlock()
	for {
		l.locker.Lock()
		if l.conn <= l.maxConn {
			l.locker.Unlock()
			return
		}
		l.locker.Unlock()
		time.Sleep(time.Duration(100) * time.Millisecond)
	}
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
	l.rps++
	if l.rpsBegin == nil {
		now := time.Now()
		l.rpsBegin = &now
		if l.rps <= l.maxRps {
			return true
		}
	}
	time.Sleep(time.Duration(10) * time.Millisecond)
	seconds := time.Since(*l.rpsBegin).Seconds()
	for float64(l.rps)/seconds > float64(l.maxRps) {
		time.Sleep(time.Duration(10) * time.Millisecond)
		seconds = time.Since(*l.rpsBegin).Seconds()
	}
	return true
}

type netConn struct {
	net.Conn
	limit *limit
}

func (c *netConn) Close() error {
	c.limit.downConn()
	log.Debugf("Close %s %s", c.Conn.RemoteAddr().Network(), c.Conn.RemoteAddr().String())
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
