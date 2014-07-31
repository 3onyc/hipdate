package main

import (
	"fmt"
	"github.com/3onyc/hipdate"
	"github.com/3onyc/hipdate/backends/hipache"
	"github.com/3onyc/hipdate/sources"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"sync"
)

type HostPortPair struct {
	hostname string
	port     uint16
}

func (pair HostPortPair) String() string {
	return fmt.Sprintf("%s:%d", pair.hostname, pair.port)
}

func main() {
	dockerUrl := os.Getenv("DOCKER_URL")
	if dockerUrl == "" {
		log.Fatalln("DOCKER_URL environment variable is not set")
	}

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatalln("REDIS_URL environment variable is not set")
	}

	d, err := docker.NewClient(dockerUrl)
	if err != nil {
		log.Fatalln("Docker:", err)
	}

	wg := &sync.WaitGroup{}
	ce := make(chan *hipdate.ChangeEvent)

	s := []hipdate.Source{
		sources.NewDockerSource(d, ce, wg),
	}

	b, err := hipache.NewHipacheBackend(redisUrl)
	if err != nil {
		log.Fatalln("ERR:", err)
	}

	app := hipdate.NewApplication(b, s, ce, wg)

	log.Println("Starting...")
	app.Start()
}
