package main

import (
	"errors"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	"log"
	"sync"

	_ "github.com/3onyc/hipdate"
)

var (
	BackendNotFoundError = errors.New("Backend not found")
)

func main() {
	cfg := LoadConfig()

	if cfg.Backend == "" {
		log.Fatalln("[FATAL] No backend selected")
	}

	if len(cfg.Sources) == 0 {
		log.Fatalln("[FATAL] No sources selected")
	}

	wg := &sync.WaitGroup{}
	ce := make(chan *shared.ChangeEvent)

	srcs := InitSources(cfg, ce, wg)
	be, err := InitBackend(cfg)
	switch {
	case err == BackendNotFoundError:
		log.Fatalf("[FATAL] Backend '%s' not found\n", cfg.Backend)
	case err != nil:
		log.Fatalf("[FATAL][backend:%s] %s", cfg.Backend, err)
	}

	app := NewApplication(be, srcs, ce, wg)

	log.Println("Starting...")
	app.Start()
}

func InitSources(
	cfg Config,
	ce chan *shared.ChangeEvent,
	wg *sync.WaitGroup,
) []sources.Source {
	srcs := []sources.Source{}

	for _, sn := range cfg.Sources {
		srcInitFn, ok := sources.SourceMap[sn]
		if !ok {
			log.Printf("[SEVERE] Source '%s' not found\n", sn)
			continue
		}

		src, err := srcInitFn(cfg.Options, ce, wg)
		if err != nil {
			log.Printf("[SEVERE][source:%s] %s", sn, err)
			continue
		}

		srcs = append(srcs, src)
	}

	return srcs
}

func InitBackend(cfg Config) (backends.Backend, error) {
	backendInitFn, ok := backends.BackendMap[cfg.Backend]
	if !ok {
		return nil, BackendNotFoundError
	}

	be, err := backendInitFn(cfg.Options)
	if err != nil {
		return nil, err
	}

	return be, nil
}
