package proxy

import (
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type proxy struct {
	*httputil.ReverseProxy
	routes  map[string]*url.URL //source to target
	origins map[string]*limit   //origin host to limit
}

func (p *proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if l, ok := p.origins[req.URL.Host]; ok {
		for !l.upRps() {
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		p.ReverseProxy.ServeHTTP(rw, req)
		return
	}

	p.ReverseProxy.ServeHTTP(rw, req)
}

func (p *proxy) dial(network, addr string) (net.Conn, error) {
	if l, ok := p.origins[addr]; ok {
		for !l.upConn() {
			log.Debugf("connection is max")
			time.Sleep(time.Duration(100) * time.Millisecond)
		}
		log.Debugf("dial %s:%s", network, addr)
		c, err := dialer.Dial(network, addr)
		return newNetConn(c, l), err
	}
	log.Infof("uncare host: %s", addr)
	c, err := dialer.Dial(network, addr)
	if err != nil {
		log.Errorf("Failed to dial %s, err: %v", addr, err)
	}
	return c, err
}

func (p *proxy) director(req *http.Request) {
	host := ""
	path := ""
	if idx := strings.Index(req.URL.Path[1:], "/"); idx != -1 {
		host = req.URL.Path[1 : idx+1]
		path = req.URL.Path[idx+1:]
	} else {
		host = req.URL.Path[1:]
	}
	log.Debugf("host: %s", host)
	if target, ok := p.routes[host]; ok {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = path
		if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}
		log.Debugf("req: %s", req.URL.String())
	}
}

func (p *proxy) addOrigin(source, target string, maxConn, rps int32) error {
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf("Cannot parse %s, err: %v", target, err)
		return err
	}
	p.routes[source] = u
	p.origins[u.Host] = &limit{
		locker:  &sync.Mutex{},
		maxConn: maxConn,
		maxRps:  rps,
		conn:    0,
		rps:     0,
	}
	return nil
}

func newProxy() *proxy {
	p := &proxy{}
	p.ReverseProxy = &httputil.ReverseProxy{}
	p.Transport = &http.Transport{
		Dial:              p.dial,
		DisableKeepAlives: true,
		Proxy:             http.ProxyFromEnvironment,
	}
	p.Director = p.director
	p.routes = make(map[string]*url.URL)
	p.origins = make(map[string]*limit)
	return p
}
