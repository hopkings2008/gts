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
	if l, ok := p.origins["www.shiqichuban.com"]; ok {
		l.upConn()
		log.Debugf("dial %s:%s", network, addr)
		c, err := dialer.Dial(network, addr)
		return newNetConn(c, l), err
	}
	log.Warnf("Uncatched host: %s", host)
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
	if target, ok := p.routes[host]; ok {
		if "limit" == host {
			//defer req.Body.Close()
			body, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return
			}
			ru := ReqUrl{}
			err = json.Unmarshal(body, &ru)
			if err != nil {
				return
			}
			u, err := url.Parse(ru.Url)
			if err != nil {
				return
			}
			req.Method = "GET"
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = u.Path
			req.Host = u.Host
			req.Header.Del("Content-Length")
			req.Header.Set("Content-Type", "")
			req.ContentLength = 0
			if l, ok := p.origins["www.shiqichuban.com"]; ok {
				l.upRps()
			}
			return
		}
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.URL.Path = path
		req.Host = target.Host
		/*if _, ok := req.Header["User-Agent"]; !ok {
			req.Header.Set("User-Agent", "")
		}*/
		//check rps
		h := getHost(req.URL.Host)
		if l, ok := p.origins[h]; ok {
			l.upRps()
		}
	}
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

func newProxy() *proxy {
	p := &proxy{}
	p.ReverseProxy = &httputil.ReverseProxy{}
	p.Transport = &http.Transport{
		Dial:                  p.dial,
		DisableKeepAlives:     false,
		Proxy:                 http.ProxyFromEnvironment,
		IdleConnTimeout:       time.Duration(5) * time.Second,
		ResponseHeaderTimeout: time.Duration(10) * time.Second,
	}
	p.Director = p.director
	p.routes = make(map[string]*url.URL)
	p.origins = make(map[string]*limit)
	p.whitelist = make(map[string]struct{})
	return p
}
