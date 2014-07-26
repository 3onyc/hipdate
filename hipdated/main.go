package main

import (
	"fmt"
	"github.com/3onyc/hipdate"
	"github.com/3onyc/hipdate/backends"
	"github.com/crosbymichael/skydock/docker"
	"github.com/garyburd/redigo/redis"
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

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		log.Fatalln("REDIS_URL environment variable is not set")
	}
	redisEndpoint, err := hipdate.ParseRedisUrl(redisUrl)
	if err != nil {
		log.Fatalln("Redis:", err)
	}
	r, err := redis.Dial("tcp", redisEndpoint)
	if err != nil {
		log.Fatalln("Redis:", err)
	}

	d, err := docker.NewClient(dockerUrl)
	if err != nil {
		log.Fatalln("Docker:", err)
	}

	b := backends.NewHipacheBackend(r)
	app := hipdate.NewApplication(b, d)

	if err := app.Initialise(); err != nil {
		log.Fatalln("Initialise:", err)
	}

	app.Watch()
}
