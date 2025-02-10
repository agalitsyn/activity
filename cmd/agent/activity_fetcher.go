package main

import (
	"fmt"
	"log"
	"strings"
	"unsafe"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/macos/foundation"
	"github.com/progrium/darwinkit/objc"

	"github.com/agalitsyn/activity/internal/model"
)

type ActivityFetcher struct{}

func NewActivityFetcher() *ActivityFetcher {
	return &ActivityFetcher{}
}

func (f *ActivityFetcher) CurrentApps() ([]model.App, error) {
	apps := []model.App{}

	objc.WithAutoreleasePool(func() {
		ws := appkit.Workspace_SharedWorkspace()
		osApps := ws.RunningApplications()
		for _, app := range osApps {
			name := app.LocalizedName()
			if name == "" {
				continue
			}
			// only UI apps
			if app.ActivationPolicy() != appkit.ApplicationActivationPolicyRegular {
				continue
			}

			var launchedAt int64
			launchDate := app.LaunchDate()
			if launchDate.Ptr() != nil {
				// Convert NSDate timeIntervalSince1970 to Go time.Time
				launchedAt = int64(launchDate.TimeIntervalSince1970())
			}

			isActive := app.IsActive()
			appData := model.App{
				Name:       name,
				IsActive:   isActive,
				LaunchedAt: launchedAt,
			}

			// Fetch additional context for active the app
			if fetchFunc, ok := AppToContextFuncsMap[name]; ok && isActive {
				data, err := fetchFunc(name)
				if err != nil {
					log.Printf("WARN failed to fetch context for %s: %v", name, err)
				}
				appData.Context = data
			}

			apps = append(apps, appData)
		}
	})

	return apps, nil
}

type ActivityContextFetchFunc func(appName string) (map[string]string, error)

var AppToContextFuncsMap = map[string]ActivityContextFetchFunc{
	"Google Chrome": BrowserContext,
	"Brave Browser": BrowserContext,
	"Firefox":       BrowserContext,
	"Safari":        BrowserContext,
}

func BrowserContext(appName string) (map[string]string, error) {
	var script string

	switch appName {
	case "Firefox":
		script = `
            tell application "Firefox"
                activate
                tell application "System Events"
                    keystroke "l" using {command down}
                    delay 0.1
                    keystroke "c" using {command down}
                end tell
                delay 0.1
                set tabTitle to the clipboard
                set tabURL to the clipboard
            end tell
            return tabTitle & "||" & tabURL
        `
	case "Safari":
		script = `
            tell application "Safari"
                set tabTitle to name of current tab of front window
                set tabURL to URL of current tab of front window
            end tell
            return tabTitle & "||" & tabURL
        `
	default: // Chrome based
		script = fmt.Sprintf(`
            tell application "%s"
                set tabTitle to get title of active tab of front window
                set tabURL to get URL of active tab of front window
            end tell
            return tabTitle & "||" & tabURL
        `, appName)
	}

	scriptObj := foundation.NewAppleScriptWithSource(script)
	if scriptObj.Ptr() == nil {
		return nil, fmt.Errorf("failed to create AppleScript object")
	}

	var errorObj objc.Object
	result := scriptObj.ExecuteAndReturnError(unsafe.Pointer(&errorObj))
	if errorObj.Ptr() != nil {
		return nil, fmt.Errorf("AppleScript error: %v", errorObj)
	}

	if result.Ptr() != nil {
		output := result.StringValue()
		parts := strings.Split(output, "||")
		if len(parts) != 2 {
			return nil, fmt.Errorf("unexpected script output: %s", output)
		}
		tabTitle := parts[0]
		// tabURL := parts[1]

		data := make(map[string]string)
		data["tabTitle"] = tabTitle
		// data["tabURL"] = tabURL
		return data, nil
	}

	return nil, nil
}
