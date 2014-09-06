package docker

import (
	"errors"
	"github.com/3onyc/hipdate/shared"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"strings"
)

var InvalidPortError = errors.New("Invalid port")

// Parse the env variable containing the hostnames
func parseHostnameVar(hostnameVar string) []string {
	return strings.Split(hostnameVar, "|")
}

func getHostnames(e docker.Env) []shared.Host {
	hosts := []shared.Host{}

	if ok := e.Exists("WEB_HOSTNAME"); ok {
		for _, host := range parseHostnameVar(e.Get("WEB_HOSTNAME")) {
			hosts = append(hosts, shared.Host(host))
		}
	}

	return hosts
}

func getPort(e docker.Env) (uint32, error) {
	if ok := e.Exists("WEB_PORT"); !ok {
		return 80, nil
	}

	p := e.GetInt("WEB_PORT")
	if p < 1 {
		return 0, InvalidPortError
	}

	return uint32(p), nil
}

func parseContainer(c *docker.Container) *ContainerData {
	env := docker.Env(c.Config.Env)
	hosts := getHostnames(env)
	port, err := getPort(env)

	if err != nil {
		log.Printf(
			"WARN Port below 0 for container %s, defaulting to 80",
			c.ID,
		)
		port = 80
	}

	return NewContainerData(
		*shared.NewEndpoint("http", c.NetworkSettings.IPAddress, port),
		hosts,
	)
}
