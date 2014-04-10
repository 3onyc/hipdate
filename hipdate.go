package main

import (
	"fmt"
	"github.com/fsouza/go-dockerclient"
	"log"
	"os"
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

	client, err := docker.NewClient(dockerUrl)
	if err != nil {
		log.Fatalln(err)
	}

	if err := initialise(client); err != nil {
		log.Fatalln(err)
	}

	if err := watch(client); err != nil {
		log.Fatalln(err)
	}
}
