package activity

import (
	"github.com/progrium/darwinkit/macos/appkit"
	"github.com/progrium/darwinkit/objc"

	"github.com/agalitsyn/activity/internal/model"
)

type Fetcher struct{}

func NewFetcher() *Fetcher {
	return &Fetcher{}
}

func (f *Fetcher) CurrentApps() ([]model.App, error) {
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

			apps = append(apps, model.App{
				Name:     name,
				IsActive: app.IsActive(),
				// LaunchedAt: foundation.DateFrom(app.LaunchDate()),
			})
		}
	})

	return apps, nil
}
