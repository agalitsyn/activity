package controller

import (
	"net/http"

	"github.com/agalitsyn/activity/cmd/server/renderer"
)

type PageController struct {
	*renderer.HTMLRenderer
}

func NewPageController(
	r *renderer.HTMLRenderer,
) *PageController {
	return &PageController{
		HTMLRenderer: r,
	}
}

func (s *PageController) HomePage(w http.ResponseWriter, r *http.Request) {
	s.Render(w, r, http.StatusOK, "home.tmpl.html", renderer.SmartBlock, nil)
}
