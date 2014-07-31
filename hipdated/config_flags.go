package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

func ConfigParseFlags() Config {
	cfg := Config{}

	backend := flag.String("backend", "", "Specify which backend to use")
	sources := flag.String("sources", "", "List of sources to use (Comma seperated)")
	options := flag.String("options", "", "Configure backend/source options (Format: \"foo=bar baz=qux\")")
	help := flag.Bool("help", false, "Print this help message")
	version := flag.Bool("version", false, "Print application version")
	flag.Parse()

	if *help {
		fmt.Fprintf(os.Stderr, "Usage: hipdated <options>\n")
		flag.PrintDefaults()
		os.Exit(0)
	}

	if *version {
		fmt.Fprintf(os.Stderr, "hipdated v%s\n", VERSION)
		os.Exit(0)
	}

	if *backend != "" {
		cfg.Backend = *backend
	}

	if *sources != "" {
		cfg.Sources = strings.Split(*sources, ",")
	}

	if *options != "" {
		cfg.Options = ParseOptions(*options)
	}

	return cfg
}
