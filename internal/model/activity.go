package model

import "time"

type App struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	// LaunchedAt time.Time `json:"launched_at"`
}

type ActivityEntry struct {
	CreatedAt time.Time `json:"created_at"`
	Apps      []App     `json:"apps"`
}
