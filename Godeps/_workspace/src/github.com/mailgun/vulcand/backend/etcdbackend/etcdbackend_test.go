// Note on debugging:
// github.com/davecgh/go-spew/spew package is extremely helpful when it comes to debugging DeepEquals issues.
// Here's how one uses it:
// spew.Printf("%#v\n vs\n %#v\n", a, b)
//
package etcdbackend

import (
	"fmt"
	"github.com/mailgun/go-etcd/etcd"
	log "github.com/mailgun/gotools-log"
	. "github.com/mailgun/vulcand/backend"
	"github.com/mailgun/vulcand/plugin/ratelimit"
	. "github.com/mailgun/vulcand/plugin/registry"
	. "gopkg.in/check.v1"
	"os"
	"strings"
	"testing"
	"time"
)

func TestEtcdBackend(t *testing.T) { TestingT(t) }

type EtcdBackendSuite struct {
	backend     *EtcdBackend
	nodes       []string
	etcdPrefix  string
	consistency string
	client      *etcd.Client
	changesC    chan interface{}
}

var _ = Suite(&EtcdBackendSuite{etcdPrefix: "/vulcandtest", consistency: etcd.STRONG_CONSISTENCY})

func (s *EtcdBackendSuite) SetUpSuite(c *C) {
	log.Init([]*log.LogConfig{&log.LogConfig{Name: "console"}})

	nodes_string := os.Getenv("VULCAND_TEST_ETCD_NODES")
	if nodes_string == "" {
		// Skips the entire suite
		c.Skip("This test requires etcd, provide comma separated nodes in VULCAND_TEST_ETCD_NODES environment variable")
		return
	}

	s.nodes = strings.Split(nodes_string, ",")
}

func (s *EtcdBackendSuite) SetUpTest(c *C) {
	// Initiate a backend with a registry
	backend, err := NewEtcdBackend(GetRegistry(), s.nodes, s.etcdPrefix, s.consistency)
	c.Assert(err, IsNil)
	s.backend = backend
	s.client = s.backend.client

	// Delete all values under the given prefix
	_, err = s.client.Get(s.etcdPrefix, false, false)
	if err != nil {
		// There's no key like this
		if !notFound(err) {
			// We haven't expected this error, oops
			c.Assert(err, IsNil)
		}
	} else {
		_, err = s.backend.client.Delete(s.etcdPrefix, true)
		c.Assert(err, IsNil)
	}

	s.changesC = make(chan interface{})
	go s.backend.WatchChanges(s.changesC, false)
}

func (s *EtcdBackendSuite) TearDownTest(c *C) {
	// Make sure we've recognized the change
	s.backend.StopWatching()
}

func (s *EtcdBackendSuite) collectChanges(c *C, expected int) []interface{} {
	changes := make([]interface{}, expected)
	for i, _ := range changes {
		select {
		case changes[i] = <-s.changesC:
			//
		case <-time.After(time.Second):
			c.Fatalf("Timeout occured")
		}
	}
	return changes
}

func (s *EtcdBackendSuite) expectChanges(c *C, expected ...interface{}) {
	changes := s.collectChanges(c, len(expected))
	for i, ch := range changes {
		c.Assert(ch, DeepEquals, expected[i])
	}
}

func (s *EtcdBackendSuite) TestAddDeleteHost(c *C) {
	host := s.makeHost("localhost")

	h, err := s.backend.AddHost(host)
	c.Assert(err, IsNil)
	c.Assert(h, Equals, host)

	s.expectChanges(c, &HostAdded{Host: host})

	err = s.backend.DeleteHost("localhost")
	c.Assert(err, IsNil)

	s.expectChanges(c, &HostDeleted{
		Name: "localhost",
	})
}

func (s *EtcdBackendSuite) TestGetters(c *C) {
	hosts, err := s.backend.GetHosts()
	c.Assert(err, IsNil)
	c.Assert(len(hosts), Equals, 0)

	upstreams, err := s.backend.GetUpstreams()
	c.Assert(err, IsNil)
	c.Assert(len(upstreams), Equals, 0)
}

// Add the host twice fails
func (s *EtcdBackendSuite) TestAddTwice(c *C) {

	_, err := s.backend.AddHost(&Host{Name: "localhost"})
	c.Assert(err, IsNil)

	_, err = s.backend.AddHost(&Host{Name: "localhost"})
	c.Assert(err, FitsTypeOf, &AlreadyExistsError{})
}

func (s *EtcdBackendSuite) TestUpstreamCRUD(c *C) {
	up := s.makeUpstream("up1", 0)
	u, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)
	c.Assert(u, Equals, up)

	s.expectChanges(c, &UpstreamAdded{Upstream: up})

	upR, err := s.backend.GetUpstream("up1")
	c.Assert(err, IsNil)
	c.Assert(upR, NotNil)
	c.Assert(upR.Id, Equals, "up1")

	err = s.backend.DeleteUpstream("up1")
	c.Assert(err, IsNil)

	s.expectChanges(c, &UpstreamDeleted{
		UpstreamId: "up1",
	})
}

func (s *EtcdBackendSuite) TestUpstreamAutoId(c *C) {
	u, err := s.backend.AddUpstream(&Upstream{Endpoints: []*Endpoint{}})

	c.Assert(err, IsNil)
	c.Assert(u, NotNil)
	s.expectChanges(c, &UpstreamAdded{Upstream: u})
}

func (s *EtcdBackendSuite) TestUpstreamTwice(c *C) {
	_, err := s.backend.AddUpstream(&Upstream{Id: "up1"})
	c.Assert(err, IsNil)

	_, err = s.backend.AddUpstream(&Upstream{Id: "up1"})
	c.Assert(err, FitsTypeOf, &AlreadyExistsError{})
}

func (s *EtcdBackendSuite) TestEndpointAddReadDelete(c *C) {
	up0 := s.makeUpstream("up1", 0)

	_, err := s.backend.AddUpstream(up0)
	c.Assert(err, IsNil)

	s.expectChanges(c, &UpstreamAdded{Upstream: up0})
	up := s.makeUpstream("up1", 1)
	e := up.Endpoints[0]

	eR, err := s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)
	c.Assert(eR, Equals, e)

	eO, err := s.backend.GetEndpoint(e.UpstreamId, e.Id)
	c.Assert(err, IsNil)
	c.Assert(eO, DeepEquals, e)

	s.expectChanges(c, &EndpointAdded{
		Upstream:          up,
		Endpoint:          e,
		AffectedLocations: []*Location{},
	})

	err = s.backend.DeleteEndpoint(up.Id, e.Id)
	c.Assert(err, IsNil)

	s.expectChanges(c, &EndpointDeleted{
		Upstream:          up0,
		EndpointId:        e.Id,
		AffectedLocations: []*Location{},
	})
}

func (s *EtcdBackendSuite) TestAddEndpointUsingSet(c *C) {
	up := s.makeUpstream("u1", 1)
	e := up.Endpoints[0]

	_, err := s.client.Set(s.backend.path("upstreams", up.Id, "endpoints", e.Id), e.Url, 0)
	c.Assert(err, IsNil)

	s.expectChanges(c, &EndpointUpdated{
		Upstream:          up,
		Endpoint:          up.Endpoints[0],
		AffectedLocations: []*Location{},
	})
}

func (s *EtcdBackendSuite) TestAddEndpointAutoId(c *C) {
	up := s.makeUpstream("up1", 1)
	e := up.Endpoints[0]
	e.Id = ""

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)
	eR, err := s.backend.AddEndpoint(e)
	c.Assert(len(eR.Id), Not(Equals), 0)
}

func (s *EtcdBackendSuite) TestDeleteBadEndpoint(c *C) {
	up := s.makeUpstream("up1", 1)

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)

	// Non existent endpoint
	c.Assert(s.backend.DeleteEndpoint(up.Id, "notHere"), FitsTypeOf, &NotFoundError{})
	// Non existent upstream
	c.Assert(s.backend.DeleteEndpoint("upNotHere", "notHere"), FitsTypeOf, &NotFoundError{})
}

func (s *EtcdBackendSuite) TestLocationAddReadDelete(c *C) {
	up := s.makeUpstream("u1", 1)
	e := up.Endpoints[0]

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)

	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)

	host := s.makeHost("localhost")

	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	s.collectChanges(c, 3)

	loc := s.makeLocation("loc1", "/hello", host, up)

	// CREATE
	locR, err := s.backend.AddLocation(loc)
	c.Assert(err, IsNil)
	c.Assert(locR, DeepEquals, loc)

	// READ
	locR2, err := s.backend.GetLocation(loc.Hostname, loc.Id)
	c.Assert(err, IsNil)
	c.Assert(locR2, DeepEquals, loc)

	s.expectChanges(c, &LocationUpstreamUpdated{
		Host:     host,
		Location: loc,
	})

	// DELETE
	c.Assert(s.backend.DeleteLocation(loc.Hostname, loc.Id), IsNil)
	s.expectChanges(c, &LocationDeleted{
		Host:       host,
		LocationId: loc.Id,
	})
}

// Make sure we can generate location id when it's not supplied
func (s *EtcdBackendSuite) TestLocationAutoId(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	e := up.Endpoints[0]

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)

	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)

	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	s.collectChanges(c, 3)

	locR, err := s.backend.AddLocation(s.makeLocation("", "/hello", host, up))
	c.Assert(err, IsNil)
	c.Assert(len(locR.Id), Not(Equals), 0)
}

func (s *EtcdBackendSuite) TestLocationUpdateUpstream(c *C) {
	up1 := s.makeUpstream("u1", 1)
	up2 := s.makeUpstream("u2", 1)

	host := s.makeHost("localhost")

	_, err := s.backend.AddUpstream(up1)
	c.Assert(err, IsNil)
	_, err = s.backend.AddEndpoint(up1.Endpoints[0])
	c.Assert(err, IsNil)

	_, err = s.backend.AddUpstream(up2)
	c.Assert(err, IsNil)
	_, err = s.backend.AddEndpoint(up2.Endpoints[0])
	c.Assert(err, IsNil)

	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	s.collectChanges(c, 5)

	loc := s.makeLocation("loc1", "/hello", host, up1)

	_, err = s.backend.AddLocation(loc)
	c.Assert(err, IsNil)
	s.collectChanges(c, 1)

	locU, err := s.backend.UpdateLocationUpstream(loc.Hostname, loc.Id, up2.Id)
	c.Assert(err, IsNil)
	c.Assert(locU.Upstream, DeepEquals, up2)

	s.expectChanges(c, &LocationUpstreamUpdated{
		Host:     host,
		Location: locU,
	})
}

func (s *EtcdBackendSuite) TestAddLocationBadUpstream(c *C) {
	host := s.makeHost("localhost")
	up1 := s.makeUpstream("u1", 1)
	loc := s.makeLocation("loc1", "/hello", host, up1)

	_, err := s.backend.AddLocation(loc)
	c.Assert(err, NotNil)
}

func (s *EtcdBackendSuite) TestAddLocationBadHost(c *C) {
	up := s.makeUpstream("u1", 1)
	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)

	host := s.makeHost("localhost")
	loc := s.makeLocation("loc1", "/hello", host, up)

	_, err = s.backend.AddLocation(loc)
	c.Assert(err, NotNil)
}

func (s *EtcdBackendSuite) TestLocationRateLimitCRUD(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	e := up.Endpoints[0]

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)
	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)
	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	s.collectChanges(c, 3)

	loc := s.makeLocation("loc1", "/hello", host, up)
	_, err = s.backend.AddLocation(loc)
	c.Assert(err, IsNil)
	s.collectChanges(c, 1)

	m := s.makeRateLimit("rl1", 10, "client.ip", 20, 1, loc)
	mR, err := s.backend.AddLocationMiddleware(loc.Hostname, loc.Id, m)
	c.Assert(mR, NotNil)
	c.Assert(err, IsNil)

	loc.Middlewares = []*MiddlewareInstance{m}
	s.expectChanges(c, &LocationMiddlewareAdded{
		Host:       host,
		Location:   loc,
		Middleware: m,
	})

	m.Middleware.(*ratelimit.RateLimit).Burst = 100
	_, err = s.backend.UpdateLocationMiddleware(loc.Hostname, loc.Id, m)
	c.Assert(err, IsNil)
	s.expectChanges(c, &LocationMiddlewareUpdated{
		Host:       host,
		Location:   loc,
		Middleware: m,
	})

	c.Assert(s.backend.DeleteLocationMiddleware(loc.Hostname, loc.Id, m.Type, m.Id), IsNil)
	loc.Middlewares = []*MiddlewareInstance{}
	s.expectChanges(c, &LocationMiddlewareDeleted{
		Host:           host,
		Location:       loc,
		MiddlewareId:   m.Id,
		MiddlewareType: m.Type,
	})
}

func (s *EtcdBackendSuite) TestLocationLimitsErrorHandling(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	loc := s.makeLocation("loc1", "/hello", host, up)

	// Location does not exist
	m := s.makeRateLimit("rl1", 10, "client.ip", 20, 1, loc)
	_, err := s.backend.AddLocationMiddleware(loc.Hostname, loc.Id, m)
	c.Assert(err, NotNil)

	_, err = s.backend.UpdateLocationMiddleware(loc.Hostname, loc.Id, m)
	c.Assert(err, NotNil)

	// Deeleteing non-existent middleware fails
	c.Assert(s.backend.DeleteLocationMiddleware(loc.Hostname, loc.Id, m.Type, m.Id), FitsTypeOf, &NotFoundError{})

	// Middleware type is not registered
	mBad := s.makeRateLimit("rl1", 10, "client.ip", 20, 1, loc)
	m.Type = "what"

	// Adding it fails
	_, err = s.backend.AddLocationMiddleware(loc.Hostname, loc.Id, mBad)
	c.Assert(err, FitsTypeOf, &NotFoundError{})

	// Updating it fails
	_, err = s.backend.UpdateLocationMiddleware(loc.Hostname, loc.Id, mBad)
	c.Assert(err, FitsTypeOf, &NotFoundError{})

	// Getting it fails
	_, err = s.backend.GetLocationMiddleware(loc.Hostname, loc.Id, mBad.Type, mBad.Id)
	c.Assert(err, FitsTypeOf, &NotFoundError{})

	// Deleting it fails
	c.Assert(s.backend.DeleteLocationMiddleware(loc.Hostname, loc.Id, "what", m.Id), FitsTypeOf, &NotFoundError{})

	// Just bad params
	_, err = s.backend.AddLocationMiddleware("", "", mBad)
	c.Assert(err, NotNil)

	// Updating it fails
	_, err = s.backend.UpdateLocationMiddleware("", "", mBad)
	c.Assert(err, NotNil)
}

func (s *EtcdBackendSuite) TestLocationMiddlewaresAutoId(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	e := up.Endpoints[0]

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)
	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)
	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	s.collectChanges(c, 3)

	loc := s.makeLocation("loc1", "/hello", host, up)
	_, err = s.backend.AddLocation(loc)
	c.Assert(err, IsNil)
	s.collectChanges(c, 1)

	m := s.makeRateLimit("", 10, "client.ip", 20, 1, loc)
	mR, err := s.backend.AddLocationMiddleware(loc.Hostname, loc.Id, m)
	c.Assert(err, IsNil)
	c.Assert(mR.Id, Not(Equals), "")
}

func (s *EtcdBackendSuite) TestGenerateChanges(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	e := up.Endpoints[0]
	loc := s.makeLocation("loc1", "/hello", host, up)
	host.Locations = []*Location{loc}
	rl := s.makeRateLimit("rl1", 10, "client.ip", 20, 1, loc)
	loc.Middlewares = []*MiddlewareInstance{rl}

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)
	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)
	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)
	_, err = s.backend.AddLocation(loc)
	c.Assert(err, IsNil)

	m := s.makeRateLimit("rl1", 10, "client.ip", 20, 1, loc)
	_, err = s.backend.AddLocationMiddleware(loc.Hostname, loc.Id, m)

	backend, err := NewEtcdBackend(GetRegistry(), s.nodes, s.etcdPrefix, s.consistency)
	c.Assert(err, IsNil)
	defer backend.StopWatching()

	s.changesC = make(chan interface{})
	go s.backend.WatchChanges(s.changesC, true)
	s.expectChanges(c,
		&UpstreamAdded{Upstream: up},
		&EndpointAdded{Upstream: up, Endpoint: e},
		&HostAdded{Host: host},
		&LocationAdded{Host: host, Location: loc},
	)
}

func (s *EtcdBackendSuite) TestDeleteUpstreamUsedByLocation(c *C) {
	up := s.makeUpstream("u1", 1)
	host := s.makeHost("localhost")
	e := up.Endpoints[0]
	loc := s.makeLocation("loc1", "/hello", host, up)

	_, err := s.backend.AddUpstream(up)
	c.Assert(err, IsNil)

	_, err = s.backend.AddEndpoint(e)
	c.Assert(err, IsNil)

	_, err = s.backend.AddHost(host)
	c.Assert(err, IsNil)

	_, err = s.backend.AddLocation(loc)
	c.Assert(err, IsNil)

	s.collectChanges(c, 4)
	c.Assert(s.backend.DeleteUpstream(up.Id), NotNil)
}

func (s *EtcdBackendSuite) makeUpstream(id string, endpoints int) *Upstream {
	up := &Upstream{
		Id:        id,
		Endpoints: []*Endpoint{},
	}

	for i := 1; i <= endpoints; i += 1 {
		e := &Endpoint{
			Id:         fmt.Sprintf("e%d", i),
			UpstreamId: up.Id,
			Url:        fmt.Sprintf("http://endpoint%d.com", i),
		}
		up.Endpoints = append(up.Endpoints, e)
	}
	return up
}

func (s *EtcdBackendSuite) makeHost(name string) *Host {
	return &Host{
		Name:      name,
		Locations: []*Location{}}
}

func (s *EtcdBackendSuite) makeLocation(id string, path string, host *Host, up *Upstream) *Location {
	return &Location{
		Id:          id,
		Hostname:    host.Name,
		Upstream:    up,
		Path:        path,
		Middlewares: []*MiddlewareInstance{},
	}
}

func (s *EtcdBackendSuite) makeRateLimit(id string, rate int, variable string, burst int, periodSeconds int, loc *Location) *MiddlewareInstance {
	rl, err := ratelimit.NewRateLimit(rate, variable, burst, periodSeconds)
	if err != nil {
		panic(err)
	}
	return &MiddlewareInstance{
		Type:       "ratelimit",
		Priority:   1,
		Id:         id,
		Middleware: rl,
	}
}
