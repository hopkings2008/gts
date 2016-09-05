package test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
	ps.origin = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	c.Assert(ps.origin, check.NotNil)
	routes := make(map[string]string)
	routes["test"] = ps.origin.URL
	p, err := proxy.NewGtsProxy(routes)
	c.Assert(err, check.IsNil)
	ps.front = httptest.NewServer(p)
	c.Assert(ps.front, check.NotNil)
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
	resp, err := http.Get(fmt.Sprintf("%s/test", ps.front.URL))
	defer resp.Body.Close()
	c.Assert(err, check.IsNil)
	c.Assert(resp.StatusCode < 300, check.Equals, true)
}
