package http

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/krtffl/torro/internal/logger"
)

// UserStatsResponse contains user voting statistics
type UserStatsResponse struct {
	UserId        string         `json:"user_id"`
	TotalVotes    int            `json:"total_votes"`
	ClassVotes    map[string]int `json:"class_votes"`
	FirstSeen     string         `json:"first_seen"`
	LastSeen      string         `json:"last_seen"`
	SnapshotCount int            `json:"snapshot_count"`
}

// handleUserStats returns statistics for the current user
func (h *Handler) handleUserStats(w http.ResponseWriter, r *http.Request) {
	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "No user session found"})
		return
	}

	// Get user from database
	user, err := h.userRepo.Get(r.Context(), userId)
	if err != nil {
		logger.Error("[User API - Stats] Couldn't get user. %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Parse ClassVotes JSONB
	var classVotes map[string]int
	if len(user.ClassVotes) > 0 {
		if err := json.Unmarshal(user.ClassVotes, &classVotes); err != nil {
			logger.Warn("[User API - Stats] Couldn't parse class votes: %v", err)
			classVotes = make(map[string]int)
		}
	} else {
		classVotes = make(map[string]int)
	}

	// Get count of user ELO snapshots
	snapshots, err := h.userEloRepo.ListByUser(r.Context(), userId)
	if err != nil {
		logger.Error("[User API - Stats] Couldn't get snapshots. %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	response := UserStatsResponse{
		UserId:        user.Id,
		TotalVotes:    user.VoteCount,
		ClassVotes:    classVotes,
		FirstSeen:     user.FirstSeen,
		LastSeen:      user.LastSeen,
		SnapshotCount: len(snapshots),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// handleUserLeaderboard returns personalized leaderboard for a class
func (h *Handler) handleUserLeaderboard(w http.ResponseWriter, r *http.Request) {
	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "No user session found"})
		return
	}

	classId := chi.URLParam(r, "classId")
	if classId == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Class ID required"})
		return
	}

	// Get personalized leaderboard
	entries, err := h.userEloRepo.GetUserLeaderboard(r.Context(), userId, classId)
	if err != nil {
		logger.Error("[User API - Leaderboard] Couldn't get leaderboard. %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Check if user has voted enough in this category
	voteCount, err := h.userRepo.GetVoteCountForClass(r.Context(), userId, classId)
	if err != nil {
		logger.Error("[User API - Leaderboard] Couldn't get vote count. %v", err)
		// Continue anyway, just log the error
		voteCount = 0
	}

	response := map[string]interface{}{
		"user_id":         userId,
		"class_id":        classId,
		"vote_count":      voteCount,
		"entries":         entries,
		"total_entries":   len(entries),
		"min_votes_met":   voteCount >= getMinVotesForClass(classId),
		"min_votes_required": getMinVotesForClass(classId),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// handleUserGlobalLeaderboard returns user's global personalized leaderboard
func (h *Handler) handleUserGlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "No user session found"})
		return
	}

	// Get personalized global leaderboard
	entries, err := h.userEloRepo.GetUserGlobalLeaderboard(r.Context(), userId)
	if err != nil {
		logger.Error("[User API - Global Leaderboard] Couldn't get leaderboard. %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}

	// Get user stats
	user, err := h.userRepo.Get(r.Context(), userId)
	if err != nil {
		logger.Error("[User API - Global Leaderboard] Couldn't get user. %v", err)
		// Continue with partial data
	}

	totalVotes := 0
	if user != nil {
		totalVotes = user.VoteCount
	}

	response := map[string]interface{}{
		"user_id":       userId,
		"total_votes":   totalVotes,
		"entries":       entries,
		"total_entries": len(entries),
		"min_votes_met": totalVotes >= 50, // Global minimum
		"min_votes_required": 50,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// getMinVotesForClass returns the minimum number of votes required to see results for a class
// These values ensure statistical significance while keeping engagement reasonable
func getMinVotesForClass(classId string) int {
	switch classId {
	case "1": // Clàssics
		return 30
	case "2": // Novetats
		return 25
	case "3": // Xocolata
		return 30
	case "4": // Albert Adrià
		return 40
	case "5": // Global (future)
		return 50
	default:
		return 25
	}
}
