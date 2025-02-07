package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/fatih/color"
	"github.com/go-pkgz/lgr"
	"github.com/gorilla/websocket"
	"github.com/natefinch/lumberjack"
	"github.com/progrium/darwinkit/macos"
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"

	"github.com/agalitsyn/activity/internal/activity"
	"github.com/agalitsyn/activity/internal/model"
	"github.com/agalitsyn/activity/version"
)

func main() {
	runtime.LockOSThread()

	// runs macOS application event loop with a callback on success
	macos.RunApp(launch)
}

func launch(app appkit.Application, delegate *appkit.ApplicationDelegate) {
	app.SetActivationPolicy(appkit.ApplicationActivationPolicyProhibited)
	app.ActivateIgnoringOtherApps(true)
	delegate.SetApplicationShouldTerminateAfterLastWindowClosed(func(appkit.Application) bool {
		return true
	})

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

	conn, _, err := websocket.DefaultDialer.Dial(cfg.ConnectionURL, nil)
	if err != nil {
		log.Fatal("could not connect to server:", err)
	}

	// close websocket on application termination
	delegate.SetApplicationWillTerminate(func(notification foundation.Notification) {
		err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("ERROR write close:", err)
		}

		conn.Close()
	})

	fetcher := activity.NewFetcher()

	// TODO: to config
	activityLogFile := &lumberjack.Logger{
		Filename:   filepath.Join("logs", "activity.log"),
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     14,
		Compress:   true,
	}
	writer := activity.NewLogActivityWriter(activityLogFile)
	defer writer.Close()

	// TODO: to config
	pollInterval := 1
	foundation.Timer_ScheduledTimerWithTimeIntervalRepeatsBlock(
		foundation.TimeInterval(pollInterval),
		true,
		func(timer foundation.Timer) {
			apps, err := fetcher.CurrentApps()
			if err != nil {
				log.Printf("ERROR failed to fetch apps: %v", err)
				return
			}

			entry := model.Message{
				CreatedAt: time.Now().Unix(),
				Apps:      apps,
			}

			b, err := json.Marshal(entry)
			if err != nil {
				log.Printf("ERROR failed to serialize activity entry: %v", err)
				return
			}
			log.Println("DEBUG", string(b))

			if err := writer.Write(b); err != nil {
				log.Printf("ERROR failed to write entry: %v", err)
			}

			if err := conn.WriteMessage(websocket.TextMessage, b); err != nil {
				log.Println("ERROR write:", err)
				return
			}
		},
	)
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
