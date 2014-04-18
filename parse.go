package main

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

// Parse the env variable containing the hostnames
func parseHostnameVar(hostnameVar string) []string {
	if !strings.Contains(hostnameVar, "|") {
		return []string{hostnameVar}
	} else {
		return strings.Split(hostnameVar, "|")
	}
}

// Parse a <hostname>:<port> pair into a HostPortPair struct
func parseHostPortPair(hostPortPair string) (*HostPortPair, error) {
	if !strings.Contains(hostPortPair, ":") {
		return nil, errors.New("No port specified in string")
	}

	pair := strings.SplitN(hostPortPair, ":", 2)
	port, err := parsePort(pair[1])
	if err != nil {
		return nil, err
	}

	return &HostPortPair{pair[0], port}, nil
}

func parsePort(portStr string) (uint16, error) {
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return 0, err
	}
	if port == 0 {
		return 0, errors.New("Port can't be 0")
	}

	return uint16(port), nil
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

func parseRedisUrl(urlStr string) (string, error) {
	redisUrl, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	if redisUrl.Scheme != "redis" {
		return "", errors.New("Scheme is not redis://")
	}

	return redisUrl.Host, nil
}
