package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/go-pkgz/lgr"

	"github.com/agalitsyn/activity/cmd/server/controller"
	"github.com/agalitsyn/activity/cmd/server/renderer"
	"github.com/agalitsyn/activity/internal/storage/mem"
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
	log.Printf("version: %s", version.String())

	if cfg.Debug {
		log.Printf("DEBUG running with config %v", cfg.String())
	}

	templates, err := renderer.NewTemplateCache(EmbedFiles, "templates", templateFuncs())
	if err != nil {
		log.Panicln("could not load templates:", err)
	}

	htmlRenderer := renderer.NewHTMLRenderer(renderer.NewTemplateRenderer(templates))
	if cfg.Debug {
		htmlRenderer.Debug = true

		log.Println("DEBUG loaded templates")
		names := make([]string, 0, len(templates))
		for name := range templates {
			names = append(names, name)
		}
		sort.Strings(names)
		fmt.Fprintln(os.Stdout, names)
	}

	clientStorage := mem.NewClientStorage()

	pageCtrl := controller.NewPageController(htmlRenderer, clientStorage)
	webSocketCtrl := controller.NewWebsocketController(clientStorage)

	h := NewRouter(
		pageCtrl,
		webSocketCtrl,
	)
	httpServer := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
	}
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(ctx, cfg.HTTP.ShutdownTimeoutSec)
		defer cancel()

		log.Println("INFO gracefully shutting down http server")
		if err := httpServer.Shutdown(shutdownCtx); err != nil {
			log.Println("ERROR shutting down http server:", err)
		}
	}()

	log.Println("INFO starting http server:", httpServer.Addr)
	if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Println("ERROR: server:", err)
	}
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
