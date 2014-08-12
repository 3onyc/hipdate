package main

import (
	"errors"
	"flag"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	_ "github.com/3onyc/hipdate"
)

const (
	VERSION = "0.1"
)

var (
	BackendNotFoundError = errors.New("Backend not found")
)

func main() {
	flag.Parse()
	cfg := LoadConfig()

	if cfg.Backend == nil {
		log.Fatalln("FATAL No backend selected")
	}

	if len(cfg.Sources) == 0 {
		log.Fatalln("FATAL No sources selected")
	}

	wg := &sync.WaitGroup{}
	ce := make(chan *shared.ChangeEvent)
	sc := make(chan bool)

	registerSignals(sc)

	srcs := InitSources(cfg, ce, wg, sc)
	if len(srcs) == 0 {
		log.Fatalf("FATAL All sources failed to initialise")
	}

	be, err := InitBackend(cfg)
	switch {
	case err == BackendNotFoundError:
		log.Fatalf("FATAL Backend '%s' not found\n", cfg.Backend.Name)
	case err != nil:
		log.Fatalf("FATAL [backend:%s] %s", cfg.Backend.Name, err)
	}

	app := NewApplication(be, srcs, ce, wg, sc)

	log.Println("NOTICE Starting...")
	app.Start()
}

func InitSources(
	cfg Config,
	ce chan *shared.ChangeEvent,
	wg *sync.WaitGroup,
	sc chan bool,
) []sources.Source {
	srcs := []sources.Source{}

	for _, s := range cfg.Sources {
		srcInitFn, ok := sources.SourceMap[s.Name]
		if !ok {
			log.Printf("ERROR Source '%s' not found\n", s.Name)
			continue
		}

		src, err := srcInitFn(s.Options, ce, wg, sc)
		if err != nil {
			log.Printf("ERROR [source:%s] %s", s.Name, err)
			continue
		}

		srcs = append(srcs, src)
	}

	return srcs
}

func InitBackend(cfg Config) (backends.Backend, error) {
	backendInitFn, ok := backends.BackendMap[cfg.Backend.Name]
	if !ok {
		return nil, BackendNotFoundError
	}

	be, err := backendInitFn(cfg.Backend.Options)
	if err != nil {
		return nil, err
	}

	return be, nil
}

func registerSignals(sc chan bool) {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		_ = <-c
		close(sc)
	}()
}
