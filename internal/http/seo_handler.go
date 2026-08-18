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
	// The wildcard group already allows everyone; the named AI stanzas are
	// documentation + insurance (an explicit record wins over any future
	// wildcard Disallow, and states policy to bots that look themselves up).
	// /api/ is crawl-control only (widget/JSON endpoints with no standalone
	// value) - nothing under it carries a noindex meta, so there's no
	// blocked+noindex conflict.
	fmt.Fprintf(w, `User-agent: *
Allow: /
Disallow: /api/

# AI search, assistant and training crawlers are explicitly welcome.
# Named groups REPLACE the wildcard group for their bot (RFC 9309), so each
# repeats the /api/ crawl-control disallow to keep policy identical.
User-agent: OAI-SearchBot
Allow: /
Disallow: /api/
User-agent: ChatGPT-User
Allow: /
Disallow: /api/
User-agent: GPTBot
Allow: /
Disallow: /api/
User-agent: ClaudeBot
Allow: /
Disallow: /api/
User-agent: Claude-SearchBot
Allow: /
Disallow: /api/
User-agent: Claude-User
Allow: /
Disallow: /api/
User-agent: PerplexityBot
Allow: /
Disallow: /api/
User-agent: Perplexity-User
Allow: /
Disallow: /api/
User-agent: Google-Extended
Allow: /
Disallow: /api/
User-agent: Applebot
Allow: /
Disallow: /api/
User-agent: Applebot-Extended
Allow: /
Disallow: /api/
User-agent: Meta-ExternalAgent
Allow: /
Disallow: /api/
User-agent: Amazonbot
Allow: /
Disallow: /api/

Sitemap: %s/sitemap.xml
`, siteBaseURL)
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
- [Rànquing de torrons](https://torro.cat/ranquing-de-torrons): the public community ranking — overall top torrons plus per-category leaders, ELO-based, with total vote counts and an updated date. Best page to cite for "which torró is best according to public votes".
- [Els millors torrons Vicens](https://torro.cat/millors-torrons-vicens): buying-guide view of the same votes — top 10 plus the leader of each category.
- [El millor torró de xocolata](https://torro.cat/millor-torro-de-xocolata): the chocolate category's full standings by votes.
- [Torrons d'Albert Adrià](https://torro.cat/torrons-albert-adria): the Adrià Natura line (Albert Adrià × Torrons Vicens) ranked by votes.
- [Ranking de turrones (español)](https://torro.cat/es/ranking-de-turrones): Spanish twin of the community ranking.
- [Turrón de Agramunt (español)](https://torro.cat/es/turron-de-agramunt): Spanish-language explainer of the Agramunt PGI and its differences with Jijona/Alicante.
- [Categories](https://torro.cat/classes): the voting categories (arenas).
- [Premsa i dades](https://torro.cat/premsa): public aggregate stats, free to cite with attribution to torro.cat.
- Product pages live at https://torro.cat/torro/{id} — one per torró, with photo, category, ELO score and ranking position.
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

	// Honest <lastmod> for the vote-driven pages only: their primary content
	// (the standings) changes when a vote lands, so the latest vote time IS
	// the content-change time. Static pages get no lastmod - a maintained-by-
	// hand date would rot and dishonest lastmod is worse than none.
	var votesLastMod string
	if latest, err := h.pressStatsRepo.LatestVoteTime(r.Context()); err != nil {
		logger.Warn("[Handler - SitemapXML] Couldn't read latest vote time. %v", err)
	} else if latest != nil {
		votesLastMod = latest.UTC().Format("2006-01-02")
	}

	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	b.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")

	staticPages := []struct {
		path     string
		priority string
		lastmod  string
	}{
		{"/", "1.0", ""},
		{"/ranquing-de-torrons", "0.9", votesLastMod},
		{"/es/ranking-de-turrones", "0.8", votesLastMod},
		{"/millors-torrons-vicens", "0.8", votesLastMod},
		{"/millor-torro-de-xocolata", "0.7", votesLastMod},
		{"/torrons-albert-adria", "0.7", votesLastMod},
		{"/classes", "0.8", ""},
		{"/premsa", "0.5", ""},
		{"/advent", "0.5", ""},
		{"/sobre", "0.6", ""},
		{"/torro-agramunt-igp", "0.6", ""},
		{"/es/turron-de-agramunt", "0.6", ""},
		{"/torro-agramunt-vs-xixona", "0.6", ""},
		{"/tipus-de-torrons", "0.6", ""},
	}
	for _, p := range staticPages {
		if p.lastmod != "" {
			fmt.Fprintf(&b, "  <url><loc>%s%s</loc><lastmod>%s</lastmod><priority>%s</priority></url>\n", siteBaseURL, p.path, p.lastmod, p.priority)
		} else {
			fmt.Fprintf(&b, "  <url><loc>%s%s</loc><priority>%s</priority></url>\n", siteBaseURL, p.path, p.priority)
		}
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
