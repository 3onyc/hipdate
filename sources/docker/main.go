package docker

import (
	"errors"
	"github.com/3onyc/hipdate/shared"
	"github.com/3onyc/hipdate/sources"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"strings"
	"sync"
)

var (
	MissingDockerUrlError = errors.New("DOCKER_URL not specified")
)

type ContainerMap map[shared.ContainerID]*ContainerData
type ContainerData struct {
	IP        shared.IPAddress
	Hostnames []shared.Host
}
type DockerSource struct {
	d          *docker.Client
	cde        chan *docker.APIEvents
	cce        chan *shared.ChangeEvent
	Containers ContainerMap
	wg         *sync.WaitGroup
}

func NewContainerData(i shared.IPAddress, h []shared.Host) *ContainerData {
	return &ContainerData{
		IP:        i,
		Hostnames: h,
	}
}

func (ds *DockerSource) eventHandler(cde chan *docker.APIEvents) {
	for {
		select {
		case e := <-cde:
			log.Printf("received (%s) %s", e.Status, e.ID)
			if err := ds.handleEvent(e); err != nil {
				log.Println(err)
			}
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
) (
	sources.Source,
	error,
) {
	du, ok := opt["DOCKER_URL"]
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
	}, nil
}

func (ds *DockerSource) Start() {
	defer ds.wg.Done()
	ds.wg.Add(1)

	ds.Initialise()

	log.Println("Starting docker event listener...")

	ds.d.AddEventListener(ds.cde)
	ds.eventHandler(ds.cde)
}

func (ds DockerSource) Stop() {
	ds.d.RemoveEventListener(ds.cde)
}

func (ds DockerSource) handleAdd(cId shared.ContainerID) error {
	c, err := ds.d.InspectContainer(string(cId))
	if err != nil {
		return err
	}

	ip := shared.IPAddress(c.NetworkSettings.IPAddress)
	hs := getHostnames(c)

	ds.Containers[cId] = NewContainerData(ip, hs)
	for _, h := range hs {
		e := shared.NewChangeEvent("add", h, ip)
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
		e := shared.NewChangeEvent("remove", h, cd.IP)
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

// Parse the env variable containing the hostnames
func parseHostnameVar(hostnameVar string) []string {
	return strings.Split(hostnameVar, "|")
}

func getHostnames(c *docker.Container) []shared.Host {
	env := docker.Env(c.Config.Env)
	hosts := []shared.Host{}

	if ok := env.Exists("WEB_HOSTNAME"); ok {
		for _, host := range parseHostnameVar(env.Get("WEB_HOSTNAME")) {
			hosts = append(hosts, shared.Host(host))
		}
	}

	return hosts
}

func init() {
	sources.SourceMap["docker"] = NewDockerSource
}
