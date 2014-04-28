package main

import (
	"github.com/crosbymichael/skydock/docker"
	"log"
)

// Initialise, adding all running containers to hipache, and removing any
// dead ones.
func (app *Application) initialise() error {
	cs, err := app.Docker.FetchAllContainers()
	if err != nil {
		return err
	}

	app.Hosts = app.gather(cs)
	app.Hosts.Initialise(app.Redis)

	return nil
}

func (app *Application) gather(cs []*docker.Container) HostList {
	hl := HostList{}
	for _, c := range cs {
		c, err := app.Docker.FetchContainer(c.Id, c.Image)
		if err != nil {
			log.Println(c.Id, err)
			continue
		}

		ip := c.NetworkSettings.IpAddress
		for _, h := range getHostnames(c) {
			hl.Add(Host(h), Backend("http://"+ip+":80"))
			app.IPs[c.Id] = ip
		}
	}
	return hl
}

func getHostnames(c *docker.Container) []Host {
	env := parseEnv(c.Config.Env)
	hosts := []Host{}

	if _, exists := env["WEB_HOSTNAME"]; exists {
		for _, host := range parseHostnameVar(env["WEB_HOSTNAME"]) {
			hosts = append(hosts, Host(host))
		}
	}

	return hosts
}
