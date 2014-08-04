package main

import (
	"github.com/3onyc/hipdate/shared"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"strings"
)

type Sources map[string]interface{}

func (s *Sources) Set(v []string) {
	*s = Sources{}
	for _, src := range v {
		(*s)[src] = nil
	}
}

type Config struct {
	Backend string
	Sources Sources
	Options shared.OptionMap
}

func NewConfig() Config {
	return Config{
		Sources: Sources{},
		Options: shared.OptionMap{},
	}
}

func (cfg *Config) Merge(cfg2 Config) {
	if cfg2.Backend != "" {
		cfg.Backend = cfg2.Backend
	}

	for src := range cfg2.Sources {
		if _, ok := cfg.Sources[src]; ok {
			continue
		}

		cfg.Sources[src] = nil
	}

	for k, v := range cfg2.Options {
		cfg.Options[k] = v
	}
}

func ParseOptions(o string) shared.OptionMap {
	opts := docker.Env(strings.Fields(o))
	return shared.OptionMap(opts.Map())
}

func LoadConfig() Config {
	log.Println("Loading config...")

	cfg := NewConfig()
	cfg.Merge(ConfigParseEnv(os.Environ()))
	cfg.Merge(ConfigParseFlags())

	return cfg
}
