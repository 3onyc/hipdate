package main

import (
	"errors"
	"log"
	"strconv"
	"strings"
)

// Parse the env variable containing the hostnames
func parseHostnameVar(hostnameVar string) []*HostPortPair {
	var hostPortPairs []string

	if !strings.Contains(hostnameVar, "|") {
		hostPortPairs = []string{hostnameVar}
	} else {
		hostPortPairs = strings.Split(hostnameVar, "|")
	}

	result := []*HostPortPair{}
	for _, hostPortPair := range hostPortPairs {
		pair, err := parseHostPortPair(hostPortPair)
		if err != nil {
			log.Println("Skipping", hostPortPair, "| Reason:", err)
			continue
		}

		result = append(result, pair)
	}

	return result
}

// Parse a <hostname>:<port> pair into a HostPortPair struct
func parseHostPortPair(hostPortPair string) (*HostPortPair, error) {
	if !strings.Contains(hostPortPair, ":") {
		return nil, errors.New("No port specified in string")
	}

	pair := strings.SplitN(hostPortPair, ":", 2)
	port, err := strconv.ParseUint(pair[1], 10, 16)
	if err != nil {
		return nil, err
	}
	if port == 0 {
		return nil, errors.New("Port can't be 0")
	}

	return &HostPortPair{pair[0], uint16(port)}, nil
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
