package http

import (
	"net/http"

	"github.com/go-chi/render"
)

// Healthcheck
type HealtcheckResponse struct {
	Answer uint `json:"answer"`
}

func (res *HealtcheckResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

func handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	render.Render(w, r, &HealtcheckResponse{42})
}
