package proxy

import (
	"net/http"

	log "github.com/Sirupsen/logrus"
)

type GtsProxy struct {
	*proxy
}

func (gts *GtsProxy) Run() {
	log.Fatalf("Failed to start GTS with error: %v", http.ListenAndServe(":9999", nil))
}

func NewGtsProxy(routes map[string]string) (*GtsProxy, error) {
	gts := &GtsProxy{newProxy()}
	for k, v := range routes {
		if err := gts.proxy.addOrigin(k, v, 0, 0); err != nil {
			log.Errorf("Failed to addOrigin(%s, %s)", k, v)
			return nil, err
		}
	}
	http.Handle("/", gts)
	return gts, nil
}
