package proxy

import (
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"

	log "github.com/Sirupsen/logrus"
)

type limit struct {
	maxConn int
	rps     int
}

type proxy struct {
	*httputil.ReverseProxy
	routes  map[string]*url.URL //source to target
	origins map[string]limit    //origin host to limit
}

func (p *proxy) dial(network, addr string) (net.Conn, error) {
	fmt.Printf("dial %s:%s\n", network, addr)
	c, err := dialer.Dial(network, addr)
	return &netConn{c}, err
}

func (p *proxy) director(req *http.Request) {
	log.Infof("ori req with host: %s, path: %s\n", req.URL.Host, req.URL.Path)
	temp := fmt.Sprintf("http:/%s", req.URL.Path)
	log.Infof("temp: %s", temp)
	if u, err := url.Parse(temp); err != nil {
		log.Infof("host: %s, path: %s", u.Host, u.Path)
		if target, ok := p.routes[u.Host]; ok {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = u.Path
			if _, ok := req.Header["User-Agent"]; !ok {
				req.Header.Set("User-Agent", "")
			}
			log.Infof("req with host: %s, path: %s\n", req.URL.Host, req.URL.Path)
		}
	}
}

func (p *proxy) addOrigin(source, target string, maxConn, rps int) error {
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf("Cannot parse %s, err: %v\n", target, err)
		return err
	}
	p.routes[source] = u
	p.origins[u.Host] = limit{
		maxConn: maxConn,
		rps:     rps,
	}
	return nil
}

func newProxy() *proxy {
	p := &proxy{}
	p.ReverseProxy = &httputil.ReverseProxy{}
	p.Transport = &http.Transport{
		Dial:              p.dial,
		DisableKeepAlives: true,
	}
	p.Director = p.director
	p.routes = make(map[string]*url.URL)
	p.origins = make(map[string]limit)
	return p
}
