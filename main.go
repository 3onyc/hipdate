package main

import (
	"github.com/fsouza/go-dockerclient"
	"log"
)

// Initialise, adding all running containers to hipache, and removing any
// dead ones.
func initialise(client *docker.Client) error {
	containers, err := client.ListContainers(docker.ListContainersOptions{})
	if err != nil {
		return err
	}

	for _, apiContainer := range containers {
		container, err := client.InspectContainer(apiContainer.ID)
		if err != nil {
			log.Println(err)
		}

		envVars := parseEnv(container.Config.Env)
		if _, ok := envVars["WEB_HOSTNAME"]; !ok {
			continue
		}

		hostPortPairs := parseHostnameVar(envVars["WEB_HOSTNAME"])

		log.Println("Container:", container.Name)
		log.Println("IP Address:", container.NetworkSettings.IPAddress)
		log.Println("Pairs:", hostPortPairs)
	}

	return nil
}

// TODO
// Watch for stop/start events on containers, removing/adding them as needed
func watch(client *docker.Client) error {
	return nil
}
