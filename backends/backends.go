package backends

import (
	"github.com/3onyc/hipdate/shared"
)

type Backend interface {
	AddEndpoint(h shared.Host, e shared.Endpoint) error
	RemoveEndpoint(h shared.Host, e shared.Endpoint) error
	ListHosts() (*shared.HostList, error)
	Initialise() error
}

type BackendInitFunc func(opt shared.OptionMap) (Backend, error)

var (
	BackendMap map[string]BackendInitFunc
)

func init() {
	BackendMap = make(map[string]BackendInitFunc)
}
