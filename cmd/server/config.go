package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/agalitsyn/flagutils"
)

const EnvPrefix = "ACTIVITY"

type Config struct {
	Debug bool

	HTTP struct {
		Addr               string
		ShutdownTimeoutSec time.Duration
	}

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

	flag.StringVar(&cfg.HTTP.Addr, "http-addr", "localhost:8080", "HTTP service address.")
	httpShutdownTimeoutSec := flag.Int("http-shutdown", 10, "HTTP service graceful shutdown timeout (sec).")

	flagutils.Prefix = EnvPrefix
	flagutils.Parse()
	flag.Parse()

	cfg.HTTP.ShutdownTimeoutSec = time.Duration(*httpShutdownTimeoutSec) * time.Second

	return cfg
}
