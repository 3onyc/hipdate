package backend

import (
	"encoding/json"
	"testing"
	"time"

	. "gopkg.in/check.v1"

	"github.com/mailgun/vulcand/plugin"
	"github.com/mailgun/vulcand/plugin/connlimit"
)

func TestBackend(t *testing.T) { TestingT(t) }

type BackendSuite struct {
}

var _ = Suite(&BackendSuite{})

func (s *BackendSuite) TestNewHost(c *C) {
	h, err := NewHost("localhost")
	c.Assert(err, IsNil)
	c.Assert(h.Name, Equals, "localhost")
	c.Assert(h.Name, Equals, h.GetId())
	c.Assert(h.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewHostBad(c *C) {
	h, err := NewHost("")
	c.Assert(err, NotNil)
	c.Assert(h, IsNil)
}

func (s *BackendSuite) TestNewLocation(c *C) {
	l, err := NewLocation("localhost", "loc1", "/home", "u1")
	c.Assert(err, IsNil)
	c.Assert(l.GetId(), Equals, "loc1")
	c.Assert(l.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewLocationWithOptions(c *C) {
	options := LocationOptions{
		Limits: LocationLimits{
			MaxMemBodyBytes: 12,
			MaxBodyBytes:    400,
		},
		FailoverPredicate:  "IsNetworkError && AttemptsLe(1)",
		Hostname:           "host1",
		TrustForwardHeader: true,
	}
	l, err := NewLocationWithOptions("localhost", "loc1", "/home", "u1", options)
	c.Assert(err, IsNil)
	c.Assert(l.GetId(), Equals, "loc1")

	o, err := l.GetOptions()
	c.Assert(err, IsNil)

	c.Assert(o.Limits.MaxMemBodyBytes, Equals, int64(12))
	c.Assert(o.Limits.MaxBodyBytes, Equals, int64(400))

	c.Assert(o.FailoverPredicate, NotNil)
	c.Assert(o.TrustForwardHeader, Equals, true)
	c.Assert(o.Hostname, Equals, "host1")
}

func (s *BackendSuite) TestNewLocationBadParams(c *C) {
	// Bad path
	_, err := NewLocation("localhost", "loc1", "** /home  -- afawf \\~", "u1")
	c.Assert(err, NotNil)

	// Empty params
	_, err = NewLocation("", "", "", "")
	c.Assert(err, NotNil)
}

func (s *BackendSuite) TestNewLocationWithBadOptions(c *C) {
	options := []LocationOptions{
		LocationOptions{
			FailoverPredicate: "bad predicate",
		},
	}
	for _, o := range options {
		l, err := NewLocationWithOptions("localhost", "loc1", "/home", "u1", o)
		c.Assert(err, NotNil)
		c.Assert(l, IsNil)
	}
}

func (s *BackendSuite) TestNewUpstream(c *C) {
	u, err := NewUpstream("u1")
	c.Assert(err, IsNil)
	c.Assert(u.GetId(), Equals, "u1")
	c.Assert(u.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewUpstreamWithOptions(c *C) {
	options := UpstreamOptions{
		Timeouts: UpstreamTimeouts{
			Read:         "1s",
			Dial:         "2s",
			TlsHandshake: "3s",
		},
		KeepAlive: UpstreamKeepAlive{
			Period:              "4s",
			MaxIdleConnsPerHost: 3,
		},
	}
	u, err := NewUpstreamWithOptions("u1", options)
	c.Assert(err, IsNil)
	c.Assert(u.GetId(), Equals, "u1")

	o, err := u.GetTransportOptions()
	c.Assert(err, IsNil)

	c.Assert(o.Timeouts.Read, Equals, time.Second)
	c.Assert(o.Timeouts.Dial, Equals, 2*time.Second)
	c.Assert(o.Timeouts.TlsHandshake, Equals, 3*time.Second)

	c.Assert(o.KeepAlive.Period, Equals, 4*time.Second)
	c.Assert(o.KeepAlive.MaxIdleConnsPerHost, Equals, 3)
}

func (s *BackendSuite) TestUpstreamOptionsEq(c *C) {
	options := []struct {
		a UpstreamOptions
		b UpstreamOptions
		e bool
	}{
		{UpstreamOptions{}, UpstreamOptions{}, true},

		{UpstreamOptions{Timeouts: UpstreamTimeouts{Dial: "1s"}}, UpstreamOptions{Timeouts: UpstreamTimeouts{Dial: "1s"}}, true},
		{UpstreamOptions{Timeouts: UpstreamTimeouts{Dial: "2s"}}, UpstreamOptions{Timeouts: UpstreamTimeouts{Dial: "1s"}}, false},
		{UpstreamOptions{Timeouts: UpstreamTimeouts{Read: "2s"}}, UpstreamOptions{Timeouts: UpstreamTimeouts{Read: "1s"}}, false},
		{UpstreamOptions{Timeouts: UpstreamTimeouts{TlsHandshake: "2s"}}, UpstreamOptions{Timeouts: UpstreamTimeouts{TlsHandshake: "1s"}}, false},

		{UpstreamOptions{KeepAlive: UpstreamKeepAlive{Period: "2s"}}, UpstreamOptions{KeepAlive: UpstreamKeepAlive{Period: "1s"}}, false},
		{UpstreamOptions{KeepAlive: UpstreamKeepAlive{MaxIdleConnsPerHost: 1}}, UpstreamOptions{KeepAlive: UpstreamKeepAlive{MaxIdleConnsPerHost: 2}}, false},
	}
	for _, o := range options {
		c.Assert(o.a.Equals(o.b), Equals, o.e)
	}
}

func (s *BackendSuite) TestNewUpstreamWithBadOptions(c *C) {
	options := []UpstreamOptions{
		UpstreamOptions{
			Timeouts: UpstreamTimeouts{
				Read: "1what?",
			},
		},
		UpstreamOptions{
			Timeouts: UpstreamTimeouts{
				Dial: "1what?",
			},
		},
		UpstreamOptions{
			Timeouts: UpstreamTimeouts{
				TlsHandshake: "1what?",
			},
		},
		UpstreamOptions{
			KeepAlive: UpstreamKeepAlive{
				Period: "1what?",
			},
		},
	}
	for _, o := range options {
		l, err := NewUpstreamWithOptions("u1", o)
		c.Assert(err, NotNil)
		c.Assert(l, IsNil)
	}
}

func (s *BackendSuite) TestNewEndpoint(c *C) {
	e, err := NewEndpoint("u1", "e1", "http://localhost")
	c.Assert(err, IsNil)
	c.Assert(e.GetId(), Equals, "e1")
	c.Assert(e.String(), Not(Equals), "")
}

func (s *BackendSuite) TestNewEndpointBadParams(c *C) {
	_, err := NewEndpoint("u1", "e1", "http---")
	c.Assert(err, NotNil)

	// Missing upstream
	_, err = NewEndpoint("", "e1", "http://localhost")
	c.Assert(err, NotNil)
}

func (s *BackendSuite) TestHostsFromJSON(c *C) {
	h, err := NewHost("localhost")
	c.Assert(err, IsNil)

	up, err := NewUpstream("up1")
	c.Assert(err, IsNil)

	e, err := NewEndpoint("u1", "e1", "http://localhost")
	c.Assert(err, IsNil)

	l, err := NewLocation("localhost", "loc1", "/path", "up1")
	c.Assert(err, IsNil)

	cl, err := connlimit.NewConnLimit(10, "client.ip")
	c.Assert(err, IsNil)

	i := &MiddlewareInstance{Id: "c1", Type: "connlimit", Middleware: cl}

	up.Endpoints = []*Endpoint{e}
	l.Upstream = up
	l.Middlewares = []*MiddlewareInstance{i}
	h.Locations = []*Location{l}
	hosts := []*Host{h}

	bytes, err := json.Marshal(map[string]interface{}{"Hosts": hosts})

	r := plugin.NewRegistry()
	c.Assert(r.AddSpec(connlimit.GetSpec()), IsNil)

	out, err := HostsFromJSON(bytes, r.GetSpec)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)
	c.Assert(out, DeepEquals, hosts)
}

func (s *BackendSuite) TestUpstreamFromJSON(c *C) {
	up, err := NewUpstream("up1")
	c.Assert(err, IsNil)

	e, err := NewEndpoint("u1", "e1", "http://localhost")
	c.Assert(err, IsNil)

	up.Endpoints = []*Endpoint{e}

	bytes, err := json.Marshal(up)
	c.Assert(err, IsNil)

	out, err := UpstreamFromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	c.Assert(out, DeepEquals, up)
}

func (s *BackendSuite) TestEndpointFromJSON(c *C) {
	e, err := NewEndpoint("u1", "e1", "http://localhost")
	c.Assert(err, IsNil)

	bytes, err := json.Marshal(e)
	c.Assert(err, IsNil)

	out, err := EndpointFromJSON(bytes)
	c.Assert(err, IsNil)
	c.Assert(out, NotNil)

	c.Assert(out, DeepEquals, e)
}
