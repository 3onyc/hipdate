package main

import (
	"fmt"
	"github.com/3onyc/hipdate/backends"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	"log"
	"os"
	"strings"
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
	opts := shared.OptionMap{}

	hb := os.Getenv("HIPDATED_BACKEND")
	if hb == "" {
		log.Fatalln("HIPDATED_BACKEND environment variable is not set")
	}

	hcsStr := os.Getenv("HIPDATED_SOURCES")
	if hb == "" {
		log.Fatalln("HIPDATED_SOURCES environment variable is not set")
	}
	hcs := strings.Split(hcsStr, ",")

	opts["DOCKER_URL"] = os.Getenv("DOCKER_URL")
	opts["REDIS_URL"] = os.Getenv("REDIS_URL")

	wg := &sync.WaitGroup{}
	ce := make(chan *shared.ChangeEvent)

	backendInitFn, ok := backends.BackendMap[hb]
	if !ok {
		log.Fatalf("ERR: Backend %s not found\n", hb)
	}

	be, err := backendInitFn(opts)
	if err != nil {
		log.Fatalln("ERR:", err)
	}

	srcs := []sources.Source{}
	for _, srcStr := range hcs {
		srcInitFn, ok := sources.SourceMap[srcStr]
		if !ok {
			log.Fatalf("ERR: Source %s not found\n", srcStr)
			continue
		}

		src, err := srcInitFn(opts, ce, wg)
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
