package hipdate

import (
	"log"
	"sync"
)

type Application struct {
	Backend     Backend
	Sources     []Source
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
		EventStream: cce,
	}
}

func (a *Application) Add(h Host, ip IPAddress) {
}

func (a *Application) Remove(h Host, ip IPAddress) {
}

func (a *Application) EventListener() {
	for ce := range a.EventStream {
		log.Printf("Event received %v\n", ce)
		u := Upstream("http://" + ce.IP + ":80")
		switch ce.Type {
		case "add":
			a.Backend.AddUpstream(ce.Host, u)
		case "remove":
			a.Backend.RemoveUpstream(ce.Host, u)
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
