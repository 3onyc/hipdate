package main

import (
	"strings"

	docker "github.com/fsouza/go-dockerclient"
)

func ConfigParseEnv(envArr []string) Config {
	cfg := NewConfig()
	env := docker.Env(envArr)

	if ok := env.Exists("HIPDATED_BACKEND"); ok {
		p := strings.SplitN(env.Get("HIPDATED_BACKEND"), ":", 2)
		if len(p) > 1 {
			opts := ParseOptions(p[1])
			cfg.Backend = NewBackend(p[0], opts)
		} else {
			cfg.Backend = NewBackend(p[0], nil)
		}
	}

	if ok := env.Exists("HIPDATED_SOURCES"); ok {
		srcs := strings.Split(env.Get("HIPDATED_SOURCES"), " ")
		for _, src := range srcs {
			p := strings.SplitN(src, ":", 2)
			if len(p) > 1 {
				opts := ParseOptions(p[1])
				cfg.Sources = append(cfg.Sources, NewSource(p[0], opts))
			} else {
				cfg.Sources = append(cfg.Sources, NewSource(p[0], nil))
			}
		}
	}

	if ok := env.Exists("HIPDATED_OPTIONS"); ok {
		cfg.Options = ParseOptions(env.Get("HIPDATED_OPTIONS"))
	}

	return cfg
}
