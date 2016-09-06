package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/zouyu/gts/proxy"
	"gopkg.in/check.v1"
)

func Test(t *testing.T) { check.TestingT(t) }

var _ = check.Suite(&proxySuite{})

type proxySuite struct {
	origin *httptest.Server
	front  *httptest.Server
}

func (ps *proxySuite) SetUpSuite(c *check.C) {
	/*ps.origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	c.Assert(ps.origin, check.NotNil)
	routes := make(map[string]string)
	routes["test"] = ps.origin.URL
	p, err := proxy.NewGtsProxy(routes)
	c.Assert(err, check.IsNil)
	ps.front = httptest.NewServer(p)
	c.Assert(ps.front, check.NotNil)*/
}

func (ps *proxySuite) TearDownSuite(c *check.C) {
	if ps.origin != nil {
		ps.origin.Close()
	}
	if ps.front != nil {
		ps.front.Close()
	}
}

func (ps *proxySuite) SetUpTest(c *check.C) {
}

func (ps *proxySuite) TearDownTest(c *check.C) {
}

func (ps *proxySuite) TestBasicProxy(c *check.C) {
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	c.Assert(origin, check.NotNil)
	defer origin.Close()
	routes := make(map[string]*proxy.TargetInfo)
	routes["test"] = &proxy.TargetInfo{
		Target: origin.URL,
	}
	p, err := proxy.NewGtsProxy(routes)
	c.Assert(err, check.IsNil)
	front := httptest.NewServer(p)
	c.Assert(front, check.NotNil)
	defer front.Close()

	resp, err := http.Get(fmt.Sprintf("%s/test", front.URL))
	defer resp.Body.Close()
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode < 300, check.Equals, true)
}

func (ps *proxySuite) TestBasicRpsFunc(c *check.C) {
	os := newOriginServer()
	origin := httptest.NewServer(os)
	c.Assert(origin, check.NotNil)
	defer origin.Close()
	routes := make(map[string]*proxy.TargetInfo)
	routes["test"] = &proxy.TargetInfo{
		Target:  origin.URL,
		MaxConn: 0,
		MaxRps:  1,
	}
	p, err := proxy.NewGtsProxy(routes)
	c.Assert(err, check.IsNil)
	front := httptest.NewServer(p)
	c.Assert(front, check.NotNil)
	defer front.Close()
	var wg sync.WaitGroup
	count := 10
	wg.Add(count)
	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()
			resp, _ := http.Get(fmt.Sprintf("%s/test", front.URL))
			if resp.Body != nil {
				resp.Body.Close()
			}
			c.Assert(err, check.IsNil)
			c.Assert(resp.StatusCode < 300, check.Equals, true)
			c.Logf("got resp %d.", resp.StatusCode)
		}()
	}
	wg.Wait()
	rps := os.rps()
	c.Assert(rps, check.Equals, 1)
}
