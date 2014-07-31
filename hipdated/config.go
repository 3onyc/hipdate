package main

import (
	"github.com/3onyc/hipdate/shared"
	docker "github.com/fsouza/go-dockerclient"
	"strings"
)

type Config struct {
	Backend string
	Sources []string
	Options shared.OptionMap
}

func ParseOptions(o string) shared.OptionMap {
	opts := docker.Env(strings.Fields(o))
	return shared.OptionMap(opts.Map())
}
