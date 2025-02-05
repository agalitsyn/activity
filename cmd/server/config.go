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

	runPrintVersion bool
	runMigrate      bool
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

	flag.BoolVar(&cfg.runPrintVersion, "version", false, "Show version.")
	flag.BoolVar(&cfg.runMigrate, "migrate", false, "Migrate.")

	flagutils.Prefix = EnvPrefix
	flagutils.Parse()
	flag.Parse()

	return cfg
}
