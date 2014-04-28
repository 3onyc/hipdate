package main

import (
	"fmt"
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
	redisEndpoint, err := parseRedisUrl(redisUrl)
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

	app := Application{
		Redis:  r,
		Docker: d,
		IPs:    IPMap{},
	}

	if err := app.initialise(); err != nil {
		log.Fatalln("Initialise:", err)
	}
	if err := app.watch(); err != nil {
		log.Fatalln("Watch:", err)
	}
}
