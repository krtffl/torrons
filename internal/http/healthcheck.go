package http

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/render"

	"github.com/krtffl/torro/internal/logger"
)

// Healthcheck
type HealtcheckResponse struct {
	Answer uint `json:"answer"`
}

func (res *HealtcheckResponse) Render(w http.ResponseWriter, r *http.Request) error {
	render.Status(r, http.StatusOK)
	return nil
}

// handleHealthcheck reports readiness. It pings the database with a short
// timeout so the container HEALTHCHECK (and any external monitor) reports
// unhealthy when the DB is unreachable, instead of returning 200 while the app
// can't actually serve requests.
func (h *Handler) handleHealthcheck(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	if err := h.db.PingContext(ctx); err != nil {
		logger.Error("[Healthcheck] Database ping failed: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = w.Write([]byte(`{"status":"unavailable"}`))
		return
	}

	render.Render(w, r, &HealtcheckResponse{42})
}
