package vulcand

import (
	"errors"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	vulcan "github.com/mailgun/vulcand/api"
	"github.com/mailgun/vulcand/plugin/registry"
	"log"
	"net/http"
	"strings"
)

var (
	MissingApiUrlError = errors.New("vulcand API endpoint not specified")
)

type VulcandBackend struct {
	v *vulcan.Client
}

func NewVulcandBackend(opts shared.OptionMap) (backends.Backend, error) {
	eu, ok := opts["vulcand"]
	if !ok {
		return nil, MissingApiUrlError
	}

	v, err := createClient(eu)
	if err != nil {
		return nil, err
	}

	return &VulcandBackend{
		v: v,
	}, nil
}

func (vb *VulcandBackend) AddUpstream(
	h shared.Host,
	u shared.Upstream,
) error {
	hName := string(h)
	uUrl := string(u)
	uId := hName + "_up"
	eId := hName + "_ep"
	lId := hName + "_loc"

	if _, err := vb.v.AddHost(string(h)); isError(err) {
		return err
	}

	if _, err := vb.v.AddUpstream(uId); isError(err) {
		return err
	}

	if _, err := vb.v.AddEndpoint(uId, eId, uUrl); isError(err) {
		return err
	}

	if _, err := vb.v.AddLocation(hName, lId, "/.*", uId); isError(err) {
		return err
	}

	return nil
}
func (vb *VulcandBackend) RemoveUpstream(
	h shared.Host,
	u shared.Upstream,
) error {
	uId := string(h) + "_up"
	eId := string(h) + "_ep"

	if _, err := vb.v.DeleteEndpoint(uId, eId); isError(err) {
		return err
	}

	return nil
}

func (vb *VulcandBackend) Initialise() error {
	hosts, err := vb.v.GetHosts()
	if err != nil {
		return err
	}

	for _, h := range hosts {
		if _, err := vb.v.DeleteHost(h.Name); isError(err) {
			return err
		}
	}

	upstreams, err := vb.v.GetUpstreams()
	if err != nil {
		return err
	}

	for _, u := range upstreams {
		if _, err := vb.v.DeleteUpstream(u.Id); isError(err) {
			return err
		}
	}

	return nil
}

func (vb *VulcandBackend) ListHosts() (*shared.HostList, error) {
	hl := shared.HostList{}

	hs, err := vb.v.GetHosts()
	if err != nil {
		return nil, err
	}

	for _, vh := range hs {
		if len(vh.Locations) == 0 {
			continue
		}

		l := vh.Locations[0]
		h := shared.Host(l.Hostname)

		hl[h] = []shared.Upstream{}

		if l.Upstream != nil && len(l.Upstream.Endpoints) > 0 {
			for _, ep := range l.Upstream.Endpoints {
				hl[h] = append(hl[h], shared.Upstream(ep.Url))
			}
		}
	}

	return &hl, nil
}

func isError(err error) bool {
	if err == nil {
		return false
	}

	log.Printf("[DEBUG][backend:vulcand] isError: %s\n", err)

	if strings.Contains(err.Error(), "already exists") {
		return false
	}

	if strings.Contains(err.Error(), "already exists") {
		return false
	}

	return true
}

func init() {
	backends.BackendMap["vulcand"] = NewVulcandBackend
}

func createClient(eu string) (*vulcan.Client, error) {
	v := vulcan.NewClient(eu, registry.GetRegistry())

	// Check if etcd is reachable
	if _, err := http.Get(eu + "/v1/"); err != nil {
		return nil, err
	}

	return v, nil
}
