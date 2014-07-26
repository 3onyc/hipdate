package sources

import (
	"github.com/3onyc/hipdate"
	"github.com/crosbymichael/skydock/docker"
	"log"
	"strings"
)

type IPMap map[hipdate.ContainerID]hipdate.IPAddress
type DockerSource struct {
	d   docker.Docker
	cce chan *hipdate.ChangeEvent
	IPs IPMap
}

func (ds *DockerSource) eventHandler(cde chan *docker.Event) {
	for e := range cde {
		log.Printf("received (%s) %s %s", e.Status, e.ContainerId, e.Image)
		if err := ds.handleEvent(e); err != nil {
			log.Println(err)
		}
	}
}

func (ds *DockerSource) handleEvent(e *docker.Event) error {
	c, err := ds.d.FetchContainer(e.ContainerId, e.Image)
	if err != nil {
		return err
	}

	for _, h := range getHostnames(c) {
		switch e.Status {
		case "die", "stop", "kill":
			ds.handleAdd(c, h)
		case "start", "restart":
			ds.handleRemove(c, h)
		}
	}

	return nil
}

func NewDockerSource(
	d docker.Docker,
	cce chan *hipdate.ChangeEvent,
) *DockerSource {
	return &DockerSource{
		d:   d,
		cce: cce,
		IPs: IPMap{},
	}
}

func (ds *DockerSource) Start() {
	ds.Initialise()
	ds.eventHandler(ds.d.GetEvents())
}

func (ds DockerSource) Stop() {

}

func (ds DockerSource) handleAdd(c *docker.Container, h hipdate.Host) {
	cId := hipdate.ContainerID(c.Id)
	ip := hipdate.IPAddress(c.NetworkSettings.IpAddress)
	ds.IPs[cId] = ip
	e := hipdate.NewChangeEvent("add", h, ip)
	ds.cce <- e
}

func (ds DockerSource) handleRemove(c *docker.Container, h hipdate.Host) {
	cId := hipdate.ContainerID(c.Id)
	ip, ok := ds.IPs[cId]
	if !ok {
		return
	}

	delete(ds.IPs, cId)
	e := hipdate.NewChangeEvent("remove", h, ip)
	ds.cce <- e
}

func (ds DockerSource) Initialise() error {
	cs, err := ds.d.FetchAllContainers()
	if err != nil {
		return err
	}

	for _, c := range cs {
		c, err := ds.d.FetchContainer(c.Id, c.Image)
		if err != nil {
			log.Println(c.Id, err)
			continue
		}

		for _, h := range getHostnames(c) {
			ds.handleAdd(c, h)
		}
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
