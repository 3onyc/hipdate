package hipdate

import (
	"log"
	"sync"
)

type Application struct {
	Backend     Backend
	Sources     []Source
	wg          *sync.WaitGroup
	EventStream chan *ChangeEvent
	sc          chan bool
}

func NewApplication(
	b Backend,
	s []Source,
	cce chan *ChangeEvent,
	wg *sync.WaitGroup,
) *Application {
	return &Application{
		Backend:     b,
		Sources:     s,
		EventStream: cce,
		wg:          wg,
	}
}

func (a *Application) Add(h Host, ip IPAddress) {
}

func (a *Application) Remove(h Host, ip IPAddress) {
}

func (a *Application) EventListener() {
	defer a.wg.Done()
	for {
		select {
		case ce := <-a.EventStream:
			log.Printf("Event received %v\n", ce)
			u := Upstream("http://" + ce.IP + ":80")
			switch ce.Type {
			case "add":
				a.Backend.AddUpstream(ce.Host, u)
			case "remove":
				a.Backend.RemoveUpstream(ce.Host, u)
			case "stop":
				break
			}
		case <-a.sc:
			return
		}
	}
}

func (a *Application) Start() {
	for _, s := range a.Sources {
		go s.Start()
	}

	log.Printf("Starting main process")

	a.wg.Add(1)
	go a.EventListener()

	a.wg.Wait()
	log.Printf("Stopping cleanly via EOF")
}
