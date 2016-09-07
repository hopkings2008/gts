package proxy

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type TargetInfo struct {
	Target  string
	MaxConn int32
	MaxRps  int32
}

type GtsProxy struct {
	*proxy
}

func (gts *GtsProxy) Run() {
	http.Handle("/", gts)
	log.Fatalf("Failed to start GTS with error: %v", http.ListenAndServe(":9999", nil))
}

func (gts *GtsProxy) AddIps(ips ...string) {
	gts.proxy.setWhiteList(ips)
}

func NewGtsProxy(routes map[string]*TargetInfo) (*GtsProxy, error) {
	gts := &GtsProxy{newProxy()}
	for k, v := range routes {
		if err := gts.proxy.addOrigin(k, v.Target, v.MaxConn, v.MaxRps); err != nil {
			log.Errorf("Failed to addOrigin(%s, %s)", k, v)
			return nil, err
		}
		log.Infof("succeed to add route: %s %s", k, v)
	}
	return gts, nil
}
