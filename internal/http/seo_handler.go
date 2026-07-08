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
// those pages, which noindex+follow deliberately allows. Deliberately allows
// AI crawlers too (GPTBot, ClaudeBot, PerplexityBot, Google-Extended, etc.
// all match "User-agent: *") - there's nothing indexed yet to protect, and
// blocking them would contradict the goal of AI-answer-engine visibility.
func robotsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nSitemap: %s/sitemap.xml\n", siteBaseURL)
}

// llmsTxt serves /llms.txt, a curated Markdown map of the site for LLM
// tools that read it (adoption is still limited industry-wide, but the cost
// of serving it is near zero). States plainly that this is an independent
// fan project, not an official Torrons Vicens property - the same framing
// as the WebSite JSON-LD on index.html and the site-wide footer disclosure,
// so no surface contradicts another about who runs this site.
func llmsTxt(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, `# Torrorèndum

> Torrorèndum is an independent fan project, not an official Torrons Vicens
> property. It lets visitors vote head-to-head on torró products, see ELO-based
> rankings, and follow a single-elimination bracket to a season champion.

## Key pages

- [Inici](https://torro.cat/): homepage, how the game works.
- [Categories](https://torro.cat/classes): the voting categories (arenas).
- [Classificació](https://torro.cat/leaderboard): live rankings (per-visitor view).
- [Premsa i dades](https://torro.cat/premsa): public aggregate stats, free to cite with attribution to torro.cat.
- [Advent](https://torro.cat/advent): daily advent-calendar duel.
- [Sobre Torrorèndum](https://torro.cat/sobre): About/FAQ - what the project is, how voting and ELO ranking work, why it isn't official.
- [IGP del Torró d'Agramunt](https://torro.cat/torro-agramunt-igp): explainer of the EU Protected Geographical Indication.
- [Torró d'Agramunt vs Torró de Xixona](https://torro.cat/torro-agramunt-vs-xixona): neutral comparison of the two regional traditions.
- [Tipus de torrons](https://torro.cat/tipus-de-torrons): glossary of common torró variety names.

## Notes for tools reading this file

- Torrons Vicens is referenced as the subject of the game, not as this
  site's publisher, author, or affiliate.
- Product names and photography mentioned across the site belong to their
  respective owners.
`)
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
		{"/sobre", "0.6"},
		{"/torro-agramunt-igp", "0.6"},
		{"/torro-agramunt-vs-xixona", "0.6"},
		{"/tipus-de-torrons", "0.6"},
	}
	for _, p := range staticPages {
		fmt.Fprintf(&b, "  <url><loc>%s%s</loc><priority>%s</priority></url>\n", siteBaseURL, p.path, p.priority)
	}

	for _, t := range torros {
		fmt.Fprintf(&b, "  <url><loc>%s/torro/%s</loc><priority>0.7</priority></url>\n", siteBaseURL, t.Id)
	}

	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - SitemapXML] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	for _, c := range classes {
		bracket, err := h.bracketRepo.GetLatestByClass(r.Context(), c.Id)
		if err != nil || bracket == nil {
			continue
		}
		fmt.Fprintf(&b, "  <url><loc>%s/bracket/%s</loc><priority>0.6</priority></url>\n", siteBaseURL, c.Id)
	}

	b.WriteString(`</urlset>`)

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	fmt.Fprint(w, b.String())
}
