package sources

import (
	"github.com/3onyc/hipdate"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"strings"
	"sync"
)

type ContainerMap map[hipdate.ContainerID]*ContainerData
type ContainerData struct {
	IP        hipdate.IPAddress
	Hostnames []hipdate.Host
}
type DockerSource struct {
	d          *docker.Client
	cde        chan *docker.APIEvents
	cce        chan *hipdate.ChangeEvent
	Containers ContainerMap
	wg         *sync.WaitGroup
}

func NewContainerData(i hipdate.IPAddress, h []hipdate.Host) *ContainerData {
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
	cId := hipdate.ContainerID(e.ID)
	switch e.Status {
	case "die", "stop", "kill":
		ds.handleRemove(cId)
	case "start", "restart":
		ds.handleAdd(cId)
	}

	return nil
}

func NewDockerSource(
	d *docker.Client,
	cce chan *hipdate.ChangeEvent,
	wg *sync.WaitGroup,
) *DockerSource {
	return &DockerSource{
		d:          d,
		cce:        cce,
		cde:        make(chan *docker.APIEvents),
		Containers: ContainerMap{},
		wg:         wg,
	}
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

func (ds DockerSource) handleAdd(cId hipdate.ContainerID) error {
	c, err := ds.d.InspectContainer(string(cId))
	if err != nil {
		return err
	}

	ip := hipdate.IPAddress(c.NetworkSettings.IPAddress)
	hs := getHostnames(c)

	ds.Containers[cId] = NewContainerData(ip, hs)
	for _, h := range hs {
		e := hipdate.NewChangeEvent("add", h, ip)
		ds.cce <- e
	}

	return nil
}

func (ds DockerSource) handleRemove(cId hipdate.ContainerID) {
	cd, ok := ds.Containers[cId]
	if !ok {
		return
	}

	delete(ds.Containers, cId)
	for _, h := range cd.Hostnames {
		e := hipdate.NewChangeEvent("remove", h, cd.IP)
		ds.cce <- e
	}
}

func (ds DockerSource) Initialise() error {
	cs, err := ds.d.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, c := range cs {
		ds.handleAdd(hipdate.ContainerID(c.ID))
	}

	return nil
}

// Parse the env variable containing the hostnames
func parseHostnameVar(hostnameVar string) []string {
	if !strings.Contains(hostnameVar, "|") {
		return []string{hostnameVar}
	} else {
		return strings.Split(hostnameVar, "|")
	}
}

// Parse the docker client env var array into a <var>:<value> map
func parseEnv(envVars []string) map[string]string {
	result := map[string]string{}

	for _, envVar := range envVars {
		pair := strings.SplitN(envVar, "=", 2)
		if len(pair) != 2 {
			continue
		} else {
			result[pair[0]] = pair[1]
		}
	}

	return result
}

func getHostnames(c *docker.Container) []hipdate.Host {
	env := parseEnv(c.Config.Env)
	hosts := []hipdate.Host{}

	if _, exists := env["WEB_HOSTNAME"]; exists {
		for _, host := range parseHostnameVar(env["WEB_HOSTNAME"]) {
			hosts = append(hosts, hipdate.Host(host))
		}
	}

	return hosts
}
