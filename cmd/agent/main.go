package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/go-pkgz/lgr"

	"github.com/agalitsyn/activity/internal/activity"
	"github.com/agalitsyn/activity/version"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := ParseFlags()
	if cfg.runPrintVersion {
		fmt.Fprintln(os.Stdout, version.String())
		os.Exit(0)
	}

	setupLogger(cfg.Debug)

	if cfg.Debug {
		log.Printf("DEBUG running with config %v", cfg.String())
	}

	log.Printf("version: %s", version.String())

	fetcher := activity.NewFetcher()

	writer := activity.NewLogActivityWriter()
	defer writer.Close()

	var wg sync.WaitGroup
	wg.Add(1)

	// TODO: to config
	ticker := time.NewTicker(1 * time.Second)
	go func() {
		defer wg.Done()
		defer ticker.Stop()

		for {
			select {
			case t := <-ticker.C:
				apps, err := fetcher.CurrentApps()
				if err != nil {
					log.Printf("ERROR failed to fetch apps: %v", err)
					continue
				}

				entry := activity.Entry{
					CreatedAt: t,
					Apps:      apps,
				}
				log.Printf("DEBUG %+v", entry)

				if err := writer.WriteEntry(entry); err != nil {
					log.Printf("ERROR failed to write entry: %v", err)
				}

			case <-ctx.Done():
				return
			}
		}
	}()

	wg.Wait()
}

func setupLogger(debug bool) {
	colorizer := lgr.Mapper{
		ErrorFunc:  func(s string) string { return color.New(color.FgHiRed).Sprint(s) },
		WarnFunc:   func(s string) string { return color.New(color.FgHiYellow).Sprint(s) },
		InfoFunc:   func(s string) string { return color.New(color.FgGreen).Sprint(s) },
		DebugFunc:  func(s string) string { return color.New(color.FgWhite).Sprint(s) },
		CallerFunc: func(s string) string { return color.New(color.FgBlue).Sprint(s) },
		TimeFunc:   func(s string) string { return color.New(color.FgCyan).Sprint(s) },
	}
	logOpts := []lgr.Option{lgr.LevelBraces, lgr.Map(colorizer)}
	if debug {
		logOpts = append(logOpts, []lgr.Option{lgr.Debug, lgr.CallerPkg, lgr.CallerFile, lgr.CallerFunc}...)
	}
	lgr.SetupStdLogger(logOpts...)
}
