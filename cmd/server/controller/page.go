package controller

import (
	"net/http"

	"github.com/agalitsyn/activity/cmd/server/renderer"
	"github.com/agalitsyn/activity/internal/model"
)

type PageController struct {
	*renderer.HTMLRenderer

	clientStorage model.ClientRepository
}

func NewPageController(
	r *renderer.HTMLRenderer,
	clientStorage model.ClientRepository,
) *PageController {
	return &PageController{
		HTMLRenderer:  r,
		clientStorage: clientStorage,
	}
}

type homePageData struct {
	Clients []model.Client
}

// @SSR
func (s *PageController) HomePage(w http.ResponseWriter, r *http.Request) {
	clients, err := s.clientStorage.FetchClients()
	if err != nil {
		http.Error(w, "failed to fetch clients", http.StatusInternalServerError)
		return
	}

	data := homePageData{Clients: clients}
	s.Render(w, r, http.StatusOK, "home.tmpl.html", renderer.SmartBlock, data)
}
