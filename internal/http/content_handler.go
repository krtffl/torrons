package http

import (
	"net/http"

	"github.com/krtffl/torro/internal/logger"
)

// about renders /sobre: the site's About/FAQ page. Static content, no DB
// dependency - same zero-repo pattern as index().
func (h *Handler) about(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - About] Incoming request")
	h.renderStaticPage(w, "about.html", isHX(r))
}

// igpExplainer renders /torro-agramunt-igp: an explainer of what "Indicació
// Geogràfica Protegida" means for Torró d'Agramunt specifically.
func (h *Handler) igpExplainer(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - IGPExplainer] Incoming request")
	h.renderStaticPage(w, "igp.html", isHX(r))
}

// agramuntVsXixona renders /torro-agramunt-vs-xixona: a neutral factual
// comparison of the two regional torró traditions/protected designations.
func (h *Handler) agramuntVsXixona(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - AgramuntVsXixona] Incoming request")
	h.renderStaticPage(w, "comparativa.html", isHX(r))
}

// torroGlossary renders /tipus-de-torrons: a glossary of common torró
// variety names and terms.
func (h *Handler) torroGlossary(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - TorroGlossary] Incoming request")
	h.renderStaticPage(w, "glossari.html", isHX(r))
}

// renderStaticPage executes a fully static template (no per-request data
// beyond the HX flag) shared by every handler in this file, following the
// same buffer/error-page convention as index() and classes().
func (h *Handler) renderStaticPage(w http.ResponseWriter, templateName string, hx bool) {
	setStaticPageCacheHeaders(w)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, templateName, Content{HX: hx}); err != nil {
		logger.Error("[Handler - StaticPage] Couldn't execute template %s. %v", templateName, err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - StaticPage] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}
