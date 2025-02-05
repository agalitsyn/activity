//go:build darwin

package activity

import (
	"runtime"

	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

type Fetcher struct{}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (f *Fetcher) CurrentApps() ([]App, error) {
	runtime.LockOSThread()

	apps := []App{}

	objc.WithAutoreleasePool(func() {
		ws := appkit.Workspace_SharedWorkspace()

		osApps := ws.RunningApplications()

		for _, app := range osApps {
			if app.LocalizedName() == "" {
				continue
			}
			if app.ActivationPolicy() != appkit.ApplicationActivationPolicyRegular {
				continue
			}

			apps = append(apps, App{
				Name:     app.LocalizedName(),
				IsActive: app.IsActive(),
				// LaunchedAt: foundation.DateFrom(app.LaunchDate()),
			})
		}
	})

	return apps, nil
}
