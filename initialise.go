package hipdate

import (
	"github.com/crosbymichael/skydock/docker"
	"log"
)

// Initialise, adding all running containers to hipache, and removing any
// dead ones.
func (app *Application) Initialise() error {
	cs, err := app.Docker.FetchAllContainers()
	if err != nil {
		return err
	}

	app.Hosts = app.gather(cs)
	app.Backend.Initialise(app.Hosts)

	return nil
}

func (app *Application) gather(cs []*docker.Container) HostList {
	hl := HostList{}
	for _, c := range cs {
		c, err := app.Docker.FetchContainer(c.Id, c.Image)
		cId := ContainerID(c.Id)
		if err != nil {
			log.Println(c.Id, err)
			continue
		}

		ip := IPAddress(c.NetworkSettings.IpAddress)
		for _, h := range getHostnames(c) {
			hl.Add(Host(h), Upstream("http://"+ip+":80"))
			app.IPs[cId] = ip
		}
	}
	return hl
}
