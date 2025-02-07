package model

type App struct {
	Name       string `json:"name"`
	IsActive   bool   `json:"is_active"`
	LaunchedAt int64  `json:"launched_at,omitempty"`
}

type ActivityEntry struct {
	CreatedAt int64 `json:"created_at"`
	Apps      []App `json:"apps"`
}
