package main

import (
	"github.com/crosbymichael/skydock/docker"
	"github.com/garyburd/redigo/redis"
	"log"
)

// Initialise, adding all running containers to hipache, and removing any
// dead ones.
func initialise(r redis.Conn, d docker.Docker) error {
	c, err := d.FetchAllContainers()
	if err != nil {
		return err
	}

	hl := gatherBackends(d, c)
	hl.Register(r)

	return nil
}

func gatherBackends(d docker.Docker, cs []*docker.Container) HostList {
	hl := HostList{}
	for _, c := range cs {
		c, err := d.FetchContainer(c.Id, c.Image)
		if err != nil {
			log.Println(c.Id, err)
			continue
		}

		ip := c.NetworkSettings.IpAddress
		env := parseEnv(c.Config.Env)
		if _, exists := env["WEB_HOSTNAME"]; !exists {
			continue
		}

		hostnames := parseHostnameVar(env["WEB_HOSTNAME"])
		for _, h := range hostnames {
			hl.Append(Host(h), Backend("http://"+ip+":80"))
		}
	}
	return hl
}
