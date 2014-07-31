package backends

import (
	"github.com/3onyc/hipdate/shared"
)

type Backend interface {
	AddUpstream(h shared.Host, u shared.Upstream) error
	RemoveUpstream(h shared.Host, u shared.Upstream) error
}

type BackendInitFunc func(opt shared.OptionMap) (Backend, error)

var (
	BackendMap map[string]BackendInitFunc
)

func init() {
	BackendMap = make(map[string]BackendInitFunc)
}
