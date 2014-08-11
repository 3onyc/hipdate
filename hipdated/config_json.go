package main

import (
	"encoding/json"
	"flag"
	"os"
)

var (
	cfgFile *string = flag.String("cfg", "", "Location of the config file")
)

func ConfigParseJson(filename string) (*Config, error) {
	var cfg *Config

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if err := json.NewDecoder(f).Decode(&cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}
