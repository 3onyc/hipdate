package docker

import (
	"errors"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"sync"
)

var (
	MissingDockerUrlError = errors.New("docker url not specified")
)

type ContainerMap map[shared.ContainerID]*ContainerData
type ContainerData struct {
	Endpoint  shared.Endpoint
	Hostnames []shared.Host
}
type DockerSource struct {
	d          *docker.Client
	cde        chan *docker.APIEvents
	cce        chan *shared.ChangeEvent
	Containers ContainerMap
	wg         *sync.WaitGroup
	sc         chan bool
}

func NewContainerData(e shared.Endpoint, h []shared.Host) *ContainerData {
	return &ContainerData{
		Endpoint:  e,
		Hostnames: h,
	}
}

func (ds *DockerSource) eventHandler(cde chan *docker.APIEvents) {
	for {
		select {
		case e := <-cde:
			log.Printf("DEBUG [source:docker] received (%s) %s", e.Status, e.ID)
			if err := ds.handleEvent(e); err != nil {
				log.Println(err)
			}
		case <-ds.sc:
			ds.Stop()
			return
		}
	}
}

func (ds *DockerSource) handleEvent(e *docker.APIEvents) error {
	cId := shared.ContainerID(e.ID)
	switch e.Status {
	case "die", "stop", "kill":
		ds.handleRemove(cId)
	case "start", "restart":
		ds.handleAdd(cId)
	}

	return nil
}

func NewDockerSource(
	opt shared.OptionMap,
	cce chan *shared.ChangeEvent,
	wg *sync.WaitGroup,
	sc chan bool,
) (
	sources.Source,
	error,
) {
	du, ok := opt["url"]
	if !ok {
		return nil, MissingDockerUrlError
	}

	d, err := docker.NewClient(du)
	if err != nil {
		return nil, err
	}

	if err := d.Ping(); err != nil {
		return nil, err
	}

	return &DockerSource{
		d:          d,
		cce:        cce,
		cde:        make(chan *docker.APIEvents),
		Containers: ContainerMap{},
		wg:         wg,
		sc:         sc,
	}, nil
}

func (ds *DockerSource) Start() {
	defer ds.wg.Done()
	ds.wg.Add(1)

	ds.Initialise()

	log.Println("NOTICE [source:docker] Starting...")

	ds.d.AddEventListener(ds.cde)
	ds.eventHandler(ds.cde)
}

func (ds DockerSource) Stop() {
	ds.d.RemoveEventListener(ds.cde)
	log.Println("NOTICE [source:docker] Stopped")
}

func (ds DockerSource) handleAdd(cId shared.ContainerID) error {
	c, err := ds.d.InspectContainer(string(cId))
	if err != nil {
		return err
	}

	cd := parseContainer(c)
	ds.Containers[cId] = cd

	for _, h := range cd.Hostnames {
		e := shared.NewChangeEvent("add", h, cd.Endpoint)
		ds.cce <- e
	}

	return nil
}

func (ds DockerSource) handleRemove(cId shared.ContainerID) {
	cd, ok := ds.Containers[cId]
	if !ok {
		return
	}

	delete(ds.Containers, cId)
	for _, h := range cd.Hostnames {
		e := shared.NewChangeEvent("remove", h, cd.Endpoint)
		ds.cce <- e
	}
}

func (ds DockerSource) Initialise() error {
	cs, err := ds.d.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, c := range cs {
		ds.handleAdd(shared.ContainerID(c.ID))
	}

	return nil
}

func init() {
	sources.SourceMap["docker"] = NewDockerSource
}
