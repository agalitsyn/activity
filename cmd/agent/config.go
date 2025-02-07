package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/agalitsyn/flagutils"
)

const EnvPrefix = "ACTIVITY"

type Config struct {
	Debug bool

	ConnectionURL string

	runPrintVersion bool
}

func (c Config) String() string {
	b, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		fmt.Fprintln(os.Stdout, err)
		os.Exit(0)
	}
	return string(b)
}

func ParseFlags() Config {
	var cfg Config

	flag.BoolVar(&cfg.Debug, "debug", false, "Debug mode.")
	flag.StringVar(&cfg.ConnectionURL, "url", "ws://localhost:8080/agent", "Connection URL.")
	flag.BoolVar(&cfg.runPrintVersion, "version", false, "Show version.")

	flagutils.Prefix = EnvPrefix
	flagutils.Parse()
	flag.Parse()

	return cfg
}
