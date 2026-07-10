package http

import (
	"net/http"
	"strconv"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

const (
	embedDefaultClassId = "5" // Global class
	embedDefaultLimit   = 10
	embedMaxLimit       = 25
)

// EmbedLeaderboardContent holds data for the embeddable leaderboard widget
// template. Unlike every other page in this app, embed_leaderboard.html is
// a fully self-contained standalone HTML document (own <style>, no header/
// topbar, no shared CSS/JS bundle) because it's meant to be loaded
// cross-origin inside a third party's <iframe>.
type EmbedLeaderboardContent struct {
	Entries []LeaderboardEntry
}

// embedLeaderboard serves a small, self-contained leaderboard widget meant
// to be embedded via <iframe> on third-party sites. It carries a footer
// attribution link back to the main site (backlink/SEO value), which is
// the whole point of shipping it.
func (h *Handler) embedLeaderboard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - EmbedLeaderboard] Incoming request")

	classId := r.URL.Query().Get("classId")
	if classId == "" {
		classId = embedDefaultClassId
	} else {
		// A specific but non-existent class must 404 rather than silently
		// falling through to an empty widget. The absent/default case
		// (embedDefaultClassId) is always valid and skips this lookup.
		classes, err := h.classRepo.List(r.Context())
		if err != nil {
			logger.Error("[Handler - EmbedLeaderboard] Couldn't list classes. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		if !classExists(classes, classId) {
			logger.Warn("[Handler - EmbedLeaderboard] Unknown classId: %s", classId)
			http.Error(w, "Class not found", http.StatusNotFound)
			return
		}
	}

	limit := embedDefaultLimit
	if raw := r.URL.Query().Get("limit"); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 {
			limit = parsed
		}
	}
	if limit > embedMaxLimit {
		limit = embedMaxLimit
	}

	// The "5"/Global class only exists as a row in the Classes table (and
	// as a Bracket.ClassId scope for Phase 2) - individual Torrons are
	// never themselves tagged with Class = "5". The established convention
	// for "the cross-category ranking" is an *empty* classId to
	// TorroRepo.ListFiltered, meaning "no class filter, all classes
	// combined" (see fetchGlobalLeaderboard in leaderboard_handler.go).
	// Translate here so the public classId=5 API (matching the class
	// picker's value everywhere else in this feature) actually returns the
	// global ranking instead of an always-empty result.
	listClassId := classId
	if classId == embedDefaultClassId {
		listClassId = ""
	}

	torrons, err := h.torroRepo.ListFiltered(r.Context(), listClassId, domain.TorroFilter{})
	if err != nil {
		logger.Error("[Handler - EmbedLeaderboard] Couldn't fetch torrons. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	entries := make([]LeaderboardEntry, 0, limit)
	rank := 1
	for _, torron := range torrons {
		// Skip discontinued products
		if torron.Discontinued {
			continue
		}

		entries = append(entries, LeaderboardEntry{
			Rank:        rank,
			TorronId:    torron.Id,
			TorronName:  torron.Name,
			TorronImage: torron.Image,
			Rating:      torron.Rating,
		})
		rank++

		if len(entries) >= limit {
			break
		}
	}

	entries = calculateRatingPercentages(entries)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "embed_leaderboard.html", EmbedLeaderboardContent{
		Entries: entries,
	}); err != nil {
		logger.Error("[Handler - EmbedLeaderboard] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Public, cacheable for a short window: this response carries no
	// per-user data (no cookie is even read - see UserMiddleware's /embed/
	// skip in middleware.go) and is safe to share across every viewer of a
	// given classId/limit combination.
	w.Header().Set("Cache-Control", "public, max-age=300")
	buf.WriteTo(w)
}
