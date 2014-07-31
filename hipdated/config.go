package main

import (
	"github.com/3onyc/hipdate/shared"
	docker "github.com/fsouza/go-dockerclient"
	"os"
	"strings"
)

type Config struct {
	Backend string
	Sources []string
	Options shared.OptionMap
}

func (cfg *Config) Merge(cfg2 Config) {
	if cfg2.Backend != "" {
		cfg.Backend = cfg2.Backend
	}

	cfg.Sources = append(cfg.Sources, cfg2.Sources...)
	for k, v := range cfg2.Options {
		cfg.Options[k] = v
	}
}

func ParseOptions(o string) shared.OptionMap {
	opts := docker.Env(strings.Fields(o))
	return shared.OptionMap(opts.Map())
}

func LoadConfig() Config {
	cfg := Config{
		Options: shared.OptionMap{},
	}

	cfg.Merge(ConfigParseEnv(os.Environ()))

	return cfg
}
