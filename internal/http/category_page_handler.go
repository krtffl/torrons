package http

import (
	"net/http"

	"github.com/krtffl/torro/internal/logger"
)

// CategoryPage describes one keyword-targeted, class-backed standings page
// ("El millor torró de xocolata segons els vots"). The ClassId values are
// the stable seed ids from migrations/000005 (+000017 rename): they are
// data, not schema, but they have never changed and every other hardcoded
// class reference in this package (getMinVotesForClass, the footer's
// /bracket/5 link) relies on the same stability.
type CategoryPage struct {
	Slug     string // URL path segment, e.g. "millor-torro-de-xocolata"
	ClassId  string
	Eyebrow  string
	Title    string // h1 (and og/twitter title)
	PageTit  string // <title>; keyword-first, may differ slightly from h1
	MetaDesc string
	Intro    string // sentence after the dynamic champion lead
}

// categoryPages are the shipped category pages. Keep slugs stable once
// indexed; add new entries rather than renaming.
var categoryPages = []CategoryPage{
	{
		Slug:     "millor-torro-de-xocolata",
		ClassId:  "3", // Xocolata
		Eyebrow:  "Xocolata",
		Title:    "El millor torró de xocolata segons els vots",
		PageTit:  "El millor torró de xocolata segons els vots — Torrorèndum",
		MetaDesc: "Quin és el millor torró de xocolata? El rànquing per vots de la categoria Xocolata del Torrorèndum: duels cara a cara entre els torrons de xocolata de Torrons Vicens, actualitzat en directe.",
		Intro:    "La categoria Xocolata del Torrorèndum enfronta els torrons de xocolata del catàleg de Torrons Vicens en duels cara a cara; aquest és el seu rànquing en directe.",
	},
	{
		Slug:     "torrons-albert-adria",
		ClassId:  "4", // Adrià Natura
		Eyebrow:  "Adrià Natura",
		Title:    "Els torrons d'Albert Adrià (Adrià Natura), per vots",
		PageTit:  "Torrons d'Albert Adrià: el rànquing de la línia Adrià Natura",
		MetaDesc: "Quin torró d'Albert Adrià val més la pena? El rànquing per vots de la línia Adrià Natura (la col·laboració d'Albert Adrià amb Torrons Vicens), votat en duels cara a cara per la comunitat.",
		Intro:    "La línia Adrià Natura és la col·laboració d'Albert Adrià amb Torrons Vicens, que porta postres icòniques d'elBulli al llenguatge del torró; aquí la comunitat n'ordena els sabors a cop de duel.",
	},
}

// CategoryPageContent is the template payload for category_ranking.html.
type CategoryPageContent struct {
	HX           bool
	Page         CategoryPage
	Entries      []LeaderboardEntry
	TotalVotes   int
	UpdatedAt    string
	UpdatedAtISO string
}

// categoryPageHandler returns the handler for one CategoryPage, rendering
// that class's standings from the shared rankingContent cache.
func (h *Handler) categoryPageHandler(page CategoryPage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger.Info("[Handler - CategoryPage] Incoming request for /%s", page.Slug)

		content, err := h.rankingContent(r)
		if err != nil {
			logger.Error("[Handler - CategoryPage] Couldn't build ranking for /%s. %v", page.Slug, err)
			h.renderErrorPage(w)
			return
		}

		var entries []LeaderboardEntry
		for _, c := range content.Categories {
			if c.Class.Id == page.ClassId {
				entries = c.Entries
				break
			}
		}

		buf := h.bpool.Get()
		defer h.bpool.Put(buf)

		if err := h.template.ExecuteTemplate(buf, "category_ranking.html", CategoryPageContent{
			HX:           isHX(r),
			Page:         page,
			Entries:      entries,
			TotalVotes:   content.TotalVotes,
			UpdatedAt:    content.UpdatedAt,
			UpdatedAtISO: content.UpdatedAtISO,
		}); err != nil {
			logger.Error("[Handler - CategoryPage] Couldn't execute template for /%s. %v", page.Slug, err)
			h.renderErrorPage(w)
			return
		}

		setStaticPageCacheHeaders(w)
		buf.WriteTo(w)
	}
}
