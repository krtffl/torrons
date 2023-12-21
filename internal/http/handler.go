package http

import (
	"html/template"
	"net/http"

	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/logger"
)

type Content struct {
	Progress uint
}

type Handler struct {
	template *template.Template
	bpool    *bpool.BufferPool
}

func NewHandler(bpool *bpool.BufferPool) *Handler {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		logger.Fatal("[Handler] - Failed to parse templates. %v", err)
	}

	return &Handler{template: tmpls, bpool: bpool}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Content -Index] Incoming request")

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "index.html", Content{
		Progress: 75,
	}); err != nil {
		logger.Error("[Handler - Index] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}
