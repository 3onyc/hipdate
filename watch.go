package main

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

func (app *Application) Add(c *docker.Container) {
	cId := ContainerID(c.Id)
	ip := IPAddress(c.NetworkSettings.IpAddress)
	b := Backend("http://" + ip + ":80")
	app.IPs[cId] = ip

	for _, h := range getHostnames(c) {
		exists, err := h.Exists(app.Redis)
		if err != nil {
			log.Println(err)
		}

		if !exists {
			if err := h.Create(app.Redis); err != nil {
				log.Println(err)
				continue
			}
		}

		if err := b.Register(app.Redis, h); err != nil {
			log.Println(err)
		}
	}
}

func (app *Application) Remove(c *docker.Container) {
	cId := ContainerID(c.Id)
	ip, ok := app.IPs[cId]
	if !ok {
		return
	}
	delete(app.IPs, cId)
	b := Backend("http://" + ip + ":80")

	for _, h := range getHostnames(c) {
		exists, err := h.Exists(app.Redis)
		if err != nil {
			log.Println(err)
		}

		if !exists {
			continue
		}

		if err := b.Unregister(app.Redis, h); err != nil {
			log.Println(err)
		}
	}
}

// TODO
// Watch for stop/start events on containers, removing/adding them as needed
func (app *Application) watch() {
	e := app.Docker.GetEvents()

	app.Status.Add(1)
	go app.eventHandler(e)

	log.Printf("Starting main process")
	app.Status.Wait()
	log.Printf("Stopping cleanly via EOF")
}
