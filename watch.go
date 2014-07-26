package hipdate

import (
	"github.com/crosbymichael/skydock/docker"
	"log"
)

func (app *Application) eventHandler(c chan *docker.Event) {
	defer app.Status.Done()

	for e := range c {
		log.Printf("received (%s) %s %s", e.Status, e.ContainerId, e.Image)

		cont, err := app.Docker.FetchContainer(e.ContainerId, e.Image)
		if err != nil {
			log.Println(e.ContainerId, err)
			continue
		}

		switch e.Status {
		case "die", "stop", "kill":
			app.Remove(cont)
		case "start", "restart":
			app.Add(cont)
		}
	}
}

func (a *Application) Add(c *docker.Container) {
	cId := ContainerID(c.Id)
	ip := IPAddress(c.NetworkSettings.IpAddress)
	u := Upstream("http://" + ip + ":80")
	a.IPs[cId] = ip

	for _, h := range getHostnames(c) {
		if err := a.Backend.AddUpstream(h, u); err != nil {
			log.Println(err)
		}
	}
}

func (a *Application) Remove(c *docker.Container) {
	cId := ContainerID(c.Id)
	ip, ok := a.IPs[cId]
	if !ok {
		return
	}
	delete(a.IPs, cId)
	u := Upstream("http://" + ip + ":80")

	for _, h := range getHostnames(c) {
		if err := a.Backend.RemoveUpstream(h, u); err != nil {
			log.Println(err)
		}
	}
}

func (app *Application) Watch() {
	e := app.Docker.GetEvents()

	app.Status.Add(1)
	go app.eventHandler(e)

	log.Printf("Starting main process")
	app.Status.Wait()
	log.Printf("Stopping cleanly via EOF")
}
