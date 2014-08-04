package sources

import (
	"github.com/3onyc/hipdate/shared"
	"sync"
)

type Source interface {
	Start()
	Stop()
}

type SourceInitFunc func(
	opt shared.OptionMap,
	cce chan *shared.ChangeEvent,
	wg *sync.WaitGroup,
	sc chan bool,
) (
	Source,
	error,
)

var (
	SourceMap map[string]SourceInitFunc
)

func init() {
	SourceMap = make(map[string]SourceInitFunc)
}
