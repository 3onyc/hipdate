package main

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

func ConfigParseEnv(envArr []string) Config {
	cfg := NewConfig()
	env := docker.Env(envArr)

	if ok := env.Exists("HIPDATED_BACKEND"); ok {
		cfg.Backend = env.Get("HIPDATED_BACKEND")
	}

	if ok := env.Exists("HIPDATED_SOURCES"); ok {
		cfg.Sources.Set(strings.Split(env.Get("HIPDATED_SOURCES"), ","))
	}

	if ok := env.Exists("HIPDATED_OPTIONS"); ok {
		cfg.Options = ParseOptions(env.Get("HIPDATED_OPTIONS"))
	}

	return cfg
}
