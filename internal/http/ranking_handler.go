package http

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// rankingTopN is how many torróns the public ranking page lists for the
// overall standings. Long enough to be a substantive, linkable resource,
// short enough to stay scannable (per-category blocks cover the tail).
const rankingTopN = 20

// rankingPerCategoryN is how many leaders each category block shows.
const rankingPerCategoryN = 3

// RankingCategory is one category block on the public ranking page: the
// class plus its current top entries.
type RankingCategory struct {
	Class   *domain.Class
	Entries []LeaderboardEntry
}

// RankingContent is the template payload for ranquing.html, the public,
// crawlable community ranking (unlike /leaderboard, which defaults to a
// per-visitor personalized view and is noindexed for that reason).
type RankingContent struct {
	HX         bool
	Entries    []LeaderboardEntry
	Categories []RankingCategory
	TotalVotes int
	// UpdatedAt is the human-readable (Catalan) date the cached standings
	// were computed, surfaced on-page as a freshness signal; UpdatedAtISO
	// is the same instant for the <time datetime> attribute.
	UpdatedAt    string
	UpdatedAtISO string
}

// rankingCacheTTL is how long a computed ranking payload is served before
// being recomputed. The standings move slowly relative to this window and
// the page is public/crawlable, so staleness is preferable to per-request
// aggregate queries (same reasoning as pressCache).
const rankingCacheTTL = 5 * time.Minute

// rankingCache memoizes the public ranking payload across requests. Same
// serve-stale-on-error pattern as pressCache: a transient DB failure serves
// the last good value rather than a 500.
var rankingCache struct {
	mu      sync.RWMutex
	content RankingContent
	expiry  time.Time
	hasData bool
}

// publicRanking handles GET /ranquing-de-torrons: the stable, indexable
// community ranking page. This page exists first for search engines and AI
// assistants ("rànquing de torrons", "quin és el millor torró") - it needs
// no cookie, shows the same content to every visitor, and links every torró
// detail page, unlike the personalized /leaderboard.
func (h *Handler) publicRanking(w http.ResponseWriter, r *http.Request) {
	h.renderRankingPage(w, r, "ranquing.html", "PublicRanking")
}

// renderRankingPage renders one of the standings-backed public pages from
// the shared rankingContent cache — the ranking-page counterpart of
// renderStaticPage. Cache headers are set only after a successful render,
// so an error response can never go out with a public max-age and get
// pinned by a shared cache.
func (h *Handler) renderRankingPage(w http.ResponseWriter, r *http.Request, templateName, logTag string) {
	logger.Info("[Handler - %s] Incoming request", logTag)

	content, err := h.rankingContent(r)
	if err != nil {
		logger.Error("[Handler - %s] Couldn't build ranking. %v", logTag, err)
		h.renderErrorPage(w)
		return
	}

	content.HX = isHX(r)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, templateName, content); err != nil {
		logger.Error("[Handler - %s] Couldn't execute template. %v", logTag, err)
		h.renderErrorPage(w)
		return
	}

	setStaticPageCacheHeaders(w)
	buf.WriteTo(w)
}

// millorsVicens handles GET /millors-torrons-vicens: a buying-guide framing
// of the same cached standings, targeting the "millors torrons Vicens" /
// "quin torró de Vicens comprar" query family (no answer page existed
// anywhere for it at the 2026-08-17 baseline). Shares rankingContent's
// cache, so it adds no query load.
func (h *Handler) millorsVicens(w http.ResponseWriter, r *http.Request) {
	h.renderRankingPage(w, r, "millors_vicens.html", "MillorsVicens")
}

// rankingContent returns the cached ranking payload, recomputing it when the
// TTL lapses. On recompute failure it serves the last good value if one
// exists, only surfacing the error when there is nothing cached at all.
func (h *Handler) rankingContent(r *http.Request) (RankingContent, error) {
	rankingCache.mu.RLock()
	if rankingCache.hasData && time.Now().Before(rankingCache.expiry) {
		content := rankingCache.content
		rankingCache.mu.RUnlock()
		return content, nil
	}
	rankingCache.mu.RUnlock()

	content, err := h.computeRankingContent(r)

	rankingCache.mu.Lock()
	defer rankingCache.mu.Unlock()

	if err != nil {
		if rankingCache.hasData {
			logger.Warn("[Handler - PublicRanking] Serving stale ranking after refresh failure. %v", err)
			return rankingCache.content, nil
		}
		return RankingContent{}, err
	}

	rankingCache.content = content
	rankingCache.expiry = time.Now().Add(rankingCacheTTL)
	rankingCache.hasData = true

	return content, nil
}

// computeRankingContent assembles the overall top-N, the per-category
// leaders, and the total vote count from the repositories.
func (h *Handler) computeRankingContent(r *http.Request) (RankingContent, error) {
	ctx := r.Context()

	global, errMsg := h.fetchGlobalLeaderboard(r, "global", domain.TorroFilter{})
	if errMsg != "" {
		return RankingContent{}, fmt.Errorf("fetching global standings: %s", errMsg)
	}
	if len(global) > rankingTopN {
		global = global[:rankingTopN]
	}
	global = calculateRatingPercentages(global)

	classes, err := h.classRepo.List(ctx)
	if err != nil {
		return RankingContent{}, fmt.Errorf("listing classes: %w", err)
	}

	categories := make([]RankingCategory, 0, len(classes))
	for _, class := range classes {
		entries, errMsg := h.fetchGlobalLeaderboard(r, class.Id, domain.TorroFilter{})
		if errMsg != "" {
			return RankingContent{}, fmt.Errorf("fetching standings for class %s: %s", class.Id, errMsg)
		}
		if len(entries) == 0 {
			continue
		}
		if len(entries) > rankingPerCategoryN {
			entries = entries[:rankingPerCategoryN]
		}
		categories = append(categories, RankingCategory{Class: class, Entries: entries})
	}

	totalVotes, err := h.pressStatsRepo.TotalVotes(ctx)
	if err != nil {
		return RankingContent{}, fmt.Errorf("counting votes: %w", err)
	}

	now := time.Now()
	return RankingContent{
		Entries:      global,
		Categories:   categories,
		TotalVotes:   totalVotes,
		UpdatedAt:    formatCatalanDate(now),
		UpdatedAtISO: now.Format("2006-01-02"),
	}, nil
}

// catalanMonths maps time.Month to lowercase Catalan month names for
// visible on-page dates.
var catalanMonths = [...]string{
	time.January:   "gener",
	time.February:  "febrer",
	time.March:     "març",
	time.April:     "abril",
	time.May:       "maig",
	time.June:      "juny",
	time.July:      "juliol",
	time.August:    "agost",
	time.September: "setembre",
	time.October:   "octubre",
	time.November:  "novembre",
	time.December:  "desembre",
}

// formatCatalanDate renders t as a Catalan long date ("17 d'agost de 2026"),
// applying the vowel elision (d'abril, d'agost, d'octubre).
func formatCatalanDate(t time.Time) string {
	month := catalanMonths[t.Month()]
	particle := "de "
	switch t.Month() {
	case time.April, time.August, time.October:
		particle = "d'"
	}
	return fmt.Sprintf("%d %s%s de %d", t.Day(), particle, month, t.Year())
}
