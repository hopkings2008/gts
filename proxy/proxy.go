package proxy

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
)

type ReqUrl struct {
	Url string `json:"url"`
}

type proxy struct {
	*httputil.ReverseProxy
	routes    map[string]*url.URL //source to target
	origins   map[string]*limit   //origin host to limit
	whitelist map[string]struct{}
	maxConn   int32
	maxRps    int32
}

func (p *proxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	if len(p.whitelist) != 0 {
		ipstr, _, _ := net.SplitHostPort(req.RemoteAddr)
		ip := net.ParseIP(ipstr)
		if _, ok := p.whitelist[string(ip)]; !ok {
			log.Warnf("ip: %s is not permitted.", req.RemoteAddr)
			rw.WriteHeader(http.StatusForbidden)
			return
		}
	}
	p.ReverseProxy.ServeHTTP(rw, req)
}

func (p *proxy) setWhiteList(ips []string) {
	for _, s := range ips {
		ip := net.ParseIP(s)
		p.whitelist[string(ip)] = struct{}{}
	}
}

func (p *proxy) dial(network, addr string) (net.Conn, error) {
	host := getHost(addr)
	if l, ok := p.origins[host]; ok {
		l.upConn()
		log.Infof("dial %s:%s", network, addr)
		c, err := dialer.Dial(network, addr)
		if err != nil {
			log.Errorf("failed to dial %s, err: %v", addr, err)
		}
		return newNetConn(c, l), err
	}
	l := &limit{
		locker:  &sync.Mutex{},
		maxConn: 0,
		maxRps:  80,
		conn:    0,
		rps:     0,
	}
	p.origins[host] = l
	l.upConn()
	log.Infof("dial %s:%s", network, addr)

	c, err := dialer.Dial(network, addr)
	if err != nil {
		log.Errorf("Failed to dial %s, err: %v", addr, err)
	}
	return newNetConn(c, l), err
}

func (p *proxy) resp(resp *http.Response) error {
	log.Infof("got resp %v", resp)
	return nil
}

func (p *proxy) director(req *http.Request) {
	host := ""
	if idx := strings.Index(req.URL.Path[1:], "/"); idx != -1 {
		host = req.URL.Path[1 : idx+1]
	} else {
		host = req.URL.Path[1:]
	}
	//defer req.Body.Close()
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		log.Errorf("got invalid req, cannot read body.")
		return
	}
	log.Infof("got req: %v, body: %s", req, string(body))
	ru := ReqUrl{}
	err = json.Unmarshal(body, &ru)
	if err != nil {
		log.Errorf("got invalid req %v", body)
		return
	}
	u, err := url.Parse(ru.Url)
	if err != nil {
		log.Errorf("failed to parse req %s", ru.Url)
		return
	}
	req.Method = "GET"
	req.URL.Scheme = u.Scheme
	req.URL.Host = u.Host
	req.URL.Path = u.Path
	req.Host = u.Host
	req.Header.Set("Content-Type", "")
	req.ContentLength = 0
	if l, ok := p.origins[host]; ok {
		l.upRps()
	} else {
		l := &limit{
			locker:  &sync.Mutex{},
			maxConn: p.maxConn,
			maxRps:  p.maxRps,
			conn:    0,
			rps:     0,
		}
		p.origins[host] = l
		l.upRps()
	}
	return
}

func (p *proxy) addOrigin(source, target string, maxConn, rps int32) error {
	u, err := url.Parse(target)
	if err != nil {
		log.Errorf("Cannot parse %s, err: %v", target, err)
		return err
	}
	p.routes[source] = u
	host := getHost(u.Host)
	p.origins[host] = &limit{
		locker:  &sync.Mutex{},
		maxConn: maxConn,
		maxRps:  rps,
		conn:    0,
		rps:     0,
	}
	return nil
}

func getHost(h string) string {
	if idx := strings.Index(h, ":"); idx != -1 {
		return h[:idx]
	}
	return h
}

func newProxy(maxConn, maxRps int32) *proxy {
	p := &proxy{}
	p.ReverseProxy = &httputil.ReverseProxy{}
	p.Transport = &http.Transport{
		Dial:                  p.dial,
		DisableKeepAlives:     false,
		Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       time.Duration(300) * time.Second,
		ResponseHeaderTimeout: time.Duration(300) * time.Second,
	}
	p.Director = p.director
	p.ModifyResponse = p.resp
	//p.ModifyResponse = p.resp
	p.routes = make(map[string]*url.URL)
	p.origins = make(map[string]*limit)
	p.whitelist = make(map[string]struct{})
	p.maxConn = maxConn
	p.maxRps = maxRps
	return p
}
