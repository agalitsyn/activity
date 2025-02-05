package activity

type App struct {
	Name     string `json:"name"`
	IsActive bool   `json:"is_active"`
	// LaunchedAt time.Time `json:"launched_at"`
}
