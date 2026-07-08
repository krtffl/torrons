package http

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/krtffl/torro/internal/logger"
)

// siteBaseURL is the canonical public origin used to build absolute URLs in
// robots.txt/sitemap.xml, matching the canonical/OG URLs hardcoded across
// public/templates/*.html.
const siteBaseURL = "https://torro.cat"

// robotsTxt serves a permissive robots.txt pointing crawlers at the sitemap.
// Pages that shouldn't be indexed (personal, randomized-per-request, or
// embed-only content) are opted out individually via a per-page
// <meta name="robots" content="noindex, ...">, not blocked here - a
// robots.txt Disallow would also stop crawlers from following links on
// those pages, which noindex+follow deliberately allows.
func robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", siteBaseURL)
}

// sitemapXML lists the stable, publicly-indexable pages: the homepage plus
// every individual torró product page. Personal/randomized/embed-only pages
// (see the noindex notes in their templates) are deliberately left out -
// listing a URL in the sitemap while marking it noindex sends crawlers a
// contradictory signal.
func (h *Handler) sitemapXML(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - SitemapXML] Incoming request")

	torros, err := h.torroRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - SitemapXML] Couldn't list torrons. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	staticPages := []struct {
		path     string
		priority string
	}{
		{"/", "1.0"},
		{"/classes", "0.8"},
		{"/premsa", "0.5"},
		{"/advent", "0.5"},
	}
	for _, p := range staticPages {
		fmt.Fprintf(&b, "  <url><loc>%s%s</loc><priority>%s</priority></url>\n", siteBaseURL, p.path, p.priority)
	}

	for _, t := range torros {
		fmt.Fprintf(&b, "  <url><loc>%s/torro/%s</loc><priority>0.7</priority></url>\n", siteBaseURL, t.Id)
	}

	b.WriteString(`</urlset>`)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprint(w, b.String())
}
