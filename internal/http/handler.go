package http

import (
	"html/template"
	"net/http"

	"github.com/go-chi/render"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

type Content struct {
	Pairings []*domain.Pairing
	Classes  []*domain.Class
	HX       bool
}

type Handler struct {
	template    *template.Template
	bpool       *bpool.BufferPool
	pairingRepo domain.PairingRepo
	torroRepo   domain.TorroRepo
	classRepo   domain.ClassRepo
}

func NewHandler(
	bpool *bpool.BufferPool,
	pairingRep domain.PairingRepo,
	torroRepo domain.TorroRepo,
	classRepo domain.ClassRepo,
) *Handler {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		logger.Fatal("[Handler] - Failed to parse templates. %v", err)
	}

	return &Handler{
		template:    tmpls,
		bpool:       bpool,
		pairingRepo: pairingRep,
		torroRepo:   torroRepo,
		classRepo:   classRepo,
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Index] Incoming request")

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "index.html", Content{}); err != nil {
		logger.Error("[Handler - Index] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func (h *Handler) classes(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Classes] Incoming request")

	classes, err := h.classRepo.List()
	if err != nil {
		logger.Error("[Handler - Classes] Couldn't list classes. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "classes.html", Content{
		Classes: classes,
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Classes ] Couldn't execute template. %v", err)
		h.template.ExecuteTemplate(w, "error.html", Content{})
		return
	}

	buf.WriteTo(w)
	return
}

func isHX(r *http.Request) bool {
	if r.Header.Get("HX-Request") == "true" {
		return true
	}
	return false
}
