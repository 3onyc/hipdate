package main

import (
	"errors"
	"net/url"
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
