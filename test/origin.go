package test

import (
	"fmt"
	"net/http"
	"sync"
	"time"
)

type originServer struct {
	locker   *sync.Mutex
	reqCount int
	start    *time.Time
}

func (os *originServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	os.locker.Lock()
	os.reqCount++
	if os.start == nil {
		now := time.Now()
		os.start = &now
	}
	os.locker.Unlock()
	fmt.Fprintf(w, "this call was relayed by the reverse proxy.\n")
}

func (os *originServer) rps() int {
	os.locker.Lock()
	defer os.locker.Unlock()
	if duration := time.Since(*os.start).Seconds(); duration != float64(0) {
		return int(float64(os.reqCount) / duration)
	} else {
		return os.reqCount
	}
}

func newOriginServer() *originServer {
	return &originServer{
		locker:   &sync.Mutex{},
		reqCount: 0,
		start:    nil,
	}
}
