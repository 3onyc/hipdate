package hipdate

import (
	"log"
	"sync"
)

type Application struct {
	Backend     Backend
	Sources     []Source
	Hosts       HostList
	Status      sync.WaitGroup
	EventStream chan *ChangeEvent
}

func NewApplication(
	b Backend,
	s []Source,
	cce chan *ChangeEvent,
) *Application {
	return &Application{
		Backend:     b,
		Sources:     s,
		Hosts:       HostList{},
		EventStream: cce,
	}
}

func (a *Application) Add(h Host, ip IPAddress) {
	u := Upstream("http://" + ip + ":80")
	a.Hosts.Add(h, u)
	a.Backend.AddUpstream(h, u)
}

func (a *Application) Remove(h Host, ip IPAddress) {
	u := Upstream("http://" + ip + ":80")
	a.Hosts.Remove(h, u)
	a.Backend.RemoveUpstream(h, u)
}

func (a *Application) EventListener() {
	for ce := range a.EventStream {
		log.Printf("Event received %v\n", ce)
		switch ce.Type {
		case "add":
			a.Add(ce.Host, ce.IP)
		case "remove":
			a.Remove(ce.Host, ce.IP)
		}
	}
}

func (a *Application) Start() {
	for _, s := range a.Sources {
		a.Status.Add(1)
		go s.Start()
	}

	log.Printf("Starting main process")
	go a.EventListener()
	a.Status.Wait()
	log.Printf("Stopping cleanly via EOF")
}
