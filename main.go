package hipdate

import (
	"github.com/crosbymichael/skydock/docker"
	"log"
	"sync"
)

type IPAddress string
type ContainerID string
type IPMap map[ContainerID]IPAddress
type Application struct {
	Backend Backend
	Docker  docker.Docker
	Hosts   HostList
	Status  sync.WaitGroup
	IPs     IPMap
}

func NewApplication(
	b Backend,
	d docker.Docker,
) *Application {
	return &Application{
		Backend: b,
		Docker:  d,
		IPs:     IPMap{},
	}
}

type Upstream string
type Host string

func (h Host) Key() string {
	return "frontend:" + string(h)
}

type UpstreamList []Upstream
type HostList map[Host]UpstreamList

func (hl HostList) Add(h Host, u Upstream) {
	hl[h] = append(hl[h], u)
}

func (hl HostList) Dump() {
	for h, ul := range hl {
		log.Println(" -", h)
		for _, u := range ul {
			log.Println("   -", u)
		}
	}
}

func getHostnames(c *docker.Container) []Host {
	env := parseEnv(c.Config.Env)
	hosts := []Host{}

	if _, exists := env["WEB_HOSTNAME"]; exists {
		for _, host := range parseHostnameVar(env["WEB_HOSTNAME"]) {
			hosts = append(hosts, Host(host))
		}
	}

	return hosts
}
