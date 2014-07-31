package main

import (
	"fmt"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	"log"
	"os"
	"sync"

	_ "github.com/3onyc/hipdate"
)

type HostPortPair struct {
	hostname string
	port     uint16
}

func (pair HostPortPair) String() string {
	return fmt.Sprintf("%s:%d", pair.hostname, pair.port)
}

func main() {
	cfg := ConfigParseEnv(os.Environ())

	if cfg.Backend == "" {
		log.Fatalln("No backend selected")
	}

	if len(cfg.Sources) == 0 {
		log.Fatalln("No sources selected")
	}

	wg := &sync.WaitGroup{}
	ce := make(chan *shared.ChangeEvent)

	backendInitFn, ok := backends.BackendMap[cfg.Backend]
	if !ok {
		log.Fatalf("ERR: Backend %s not found\n", cfg.Backend)
	}

	be, err := backendInitFn(cfg.Options)
	if err != nil {
		log.Fatalln("ERR:", err)
	}

	srcs := []sources.Source{}
	for _, sn := range cfg.Sources {
		srcInitFn, ok := sources.SourceMap[sn]
		if !ok {
			log.Fatalf("ERR: Source %s not found\n", sn)
			continue
		}

		src, err := srcInitFn(cfg.Options, ce, wg)
		if err != nil {
			log.Fatalln("ERR", err)
			continue
		}

		srcs = append(srcs, src)
	}

	app := NewApplication(be, srcs, ce, wg)

	log.Println("Starting...")
	app.Start()
}
