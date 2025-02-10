package model

type App struct {
	Name       string            `json:"name"`
	IsActive   bool              `json:"is_active"`
	LaunchedAt int64             `json:"launched_at,omitempty"`
	Context    map[string]string `json:"context,omitempty"`
}

type Message struct {
	CreatedAt int64 `json:"created_at"`
	Apps      []App `json:"apps"`
}
