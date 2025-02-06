package activity

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"
)

type App struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	// LaunchedAt time.Time `json:"launched_at"`
}

type Fetcher struct{}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (f *Fetcher) CurrentApps() ([]App, error) {
	apps := []App{}

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

			apps = append(apps, App{
				Name:     name,
				IsActive: app.IsActive(),
				// LaunchedAt: foundation.DateFrom(app.LaunchDate()),
			})
		}
	})

	return apps, nil
}
