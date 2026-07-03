package http

import (
	"net/http"

	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/internal/sharecard"
)

// shareCard renders and serves a 1080x1920 PNG "story" card summarizing
// the current (anonymous, cookie-identified) user's voting result: their
// top-rated torró from their personal ELO leaderboard, its rank among the
// torrons they've rated, and their total vote count.
//
// If the user hasn't voted yet, a fallback "vota per generar la teva
// targeta" card is rendered instead of erroring, matching this handler's
// siblings (stats, history).
//
// The actual drawing lives in internal/sharecard, which is a pure
// data-in/PNG-out package with no knowledge of HTTP or the database, so
// it can be reused by future features (e.g. the roadmap's "Wrapped"
// reveal) without touching this handler.
func (h *Handler) shareCard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - ShareCard] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - ShareCard] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	user, err := h.userRepo.Get(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - ShareCard] Couldn't get user. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	entries, err := h.userEloRepo.GetUserGlobalLeaderboard(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - ShareCard] Couldn't get user leaderboard. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var data sharecard.Data
	if user.VoteCount > 0 && len(entries) > 0 {
		top := entries[0]
		data = sharecard.Data{
			HasVotes:         true,
			TotalVotes:       user.VoteCount,
			TopTorroName:     top.TorronName,
			TopTorroRank:     top.Rank,
			RatedTorronCount: len(entries),
		}
	}

	png, err := sharecard.Render(data)
	if err != nil {
		logger.Error("[Handler - ShareCard] Couldn't render card. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Personalized per-user content: never cache/share across users.
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		logger.Error("[Handler - ShareCard] Couldn't write response. %v", err)
	}
}
