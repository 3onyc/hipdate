package main

import (
	"github.com/3onyc/hipdate/shared"
	docker "github.com/fsouza/go-dockerclient"
	"log"
	"os"
	"strings"
)

type Source struct {
	Name    string
	Options shared.OptionMap
}

func NewSource(n string, o shared.OptionMap) *Source {
	if o == nil {
		o = shared.OptionMap{}
	}

	return &Source{n, o}
}

type Backend struct {
	Name    string
	Options shared.OptionMap
}

func NewBackend(n string, o shared.OptionMap) *Backend {
	if o == nil {
		o = shared.OptionMap{}
	}

	return &Backend{n, o}
}

type Config struct {
	Backend *Backend
	Sources []*Source
	Options shared.OptionMap
}

func NewConfig() Config {
	return Config{
		Sources: []*Source{},
		Options: shared.OptionMap{},
	}
}

func (cfg *Config) Merge(cfg2 Config) {
	if cfg2.Backend != nil {
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
	log.Println("NOTICE Loading config...")

	cfg := NewConfig()
	cfg.Merge(ConfigParseEnv(os.Environ()))

	if *cfgFile != "" {
		cfg2, err := ConfigParseJson(*cfgFile)
		if err != nil {
			log.Printf("ERROR Failed to load %s: %s\n", *cfgFile, err)
		} else {
			cfg.Merge(*cfg2)
		}
	}

	return cfg
}
