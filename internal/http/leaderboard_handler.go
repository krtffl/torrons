package http

import (
	"fmt"
	"math"
	"net/http"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// LeaderboardEntry represents a single entry in the leaderboard display
type LeaderboardEntry struct {
	Rank             int     `json:"rank"`
	TorronId         string  `json:"torron_id"`
	TorronName       string  `json:"torron_name"`
	TorronImage      string  `json:"torron_image"`
	Rating           float64 `json:"rating"`
	VoteCount        int     `json:"vote_count"`
	RatingPercentage int     `json:"rating_percentage"` // For visual bar (0-100)
}

// LeaderboardContent holds data for template rendering
type LeaderboardContent struct {
	HX                  bool
	Title               string
	ViewType            string // "personal" or "global"
	SelectedCategory    string
	ShowCategoryFilter  bool
	Categories          []*domain.Class
	Entries             []LeaderboardEntry
	Error               string
	MinVotes            int
}

// leaderboard handles the main leaderboard page with view and category selection
func (h *Handler) leaderboard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Leaderboard] Incoming request")

	// Get query parameters
	viewType := r.URL.Query().Get("view")
	if viewType == "" {
		viewType = "personal" // Default to personal view
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = "global" // Default to global category
	}

	userId := getUserId(r)
	if userId == "" {
		logger.Error("[Handler - Leaderboard] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get all classes for category selector
	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - Leaderboard] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var entries []LeaderboardEntry
	var errorMsg string
	var minVotes int
	var title string

	// Fetch data based on view type
	if viewType == "personal" {
		entries, errorMsg, minVotes = h.fetchPersonalLeaderboard(r, userId, category)
		if category == "global" {
			title = "Els meus resultats - Global"
		} else {
			className := h.getClassName(classes, category)
			title = fmt.Sprintf("Els meus resultats - %s", className)
		}
	} else {
		entries, errorMsg = h.fetchGlobalLeaderboard(r, category)
		if category == "global" {
			title = "Resultats globals - Millor torró absolut"
		} else {
			className := h.getClassName(classes, category)
			title = fmt.Sprintf("Resultats globals - %s", className)
		}
	}

	// Calculate rating percentages for visual bars
	if len(entries) > 0 {
		maxRating := entries[0].Rating // Assuming sorted by rating desc
		minRating := entries[len(entries)-1].Rating
		ratingRange := maxRating - minRating

		for i := range entries {
			if ratingRange > 0 {
				normalized := (entries[i].Rating - minRating) / ratingRange
				entries[i].RatingPercentage = int(math.Max(10, normalized*100)) // Min 10% for visibility
			} else {
				entries[i].RatingPercentage = 100 // All same rating
			}
		}
	}

	content := LeaderboardContent{
		HX:                 isHX(r),
		Title:              title,
		ViewType:           viewType,
		SelectedCategory:   category,
		ShowCategoryFilter: true,
		Categories:         classes,
		Entries:            entries,
		Error:              errorMsg,
		MinVotes:           minVotes,
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "leaderboard.html", content); err != nil {
		logger.Error("[Handler - Leaderboard] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// fetchPersonalLeaderboard gets personalized rankings for a user
func (h *Handler) fetchPersonalLeaderboard(r *http.Request, userId, category string) ([]LeaderboardEntry, string, int) {
	minVotes := getMinVotesForClass(category)

	// Check if user has enough votes
	var voteCount int
	var err error
	if category == "global" {
		// For global, check total vote count
		user, err := h.userRepo.Get(r.Context(), userId)
		if err != nil {
			logger.Error("[Handler - Leaderboard] Couldn't get user. %v", err)
			return nil, "Error al carregar els resultats", 0
		}
		voteCount = user.VoteCount
	} else {
		// For specific class, check class vote count
		voteCount, err = h.userRepo.GetVoteCountForClass(r.Context(), userId, category)
		if err != nil {
			logger.Error("[Handler - Leaderboard] Couldn't get vote count. %v", err)
			return nil, "Error al carregar els resultats", 0
		}
	}

	if voteCount < minVotes {
		return nil, fmt.Sprintf("No tens prou vots per veure els resultats personalitzats"), minVotes
	}

	// Fetch personalized leaderboard
	var apiEntries []domain.UserLeaderboardEntry
	if category == "global" {
		apiEntries, err = h.userEloRepo.GetUserGlobalLeaderboard(r.Context(), userId, 100)
	} else {
		apiEntries, err = h.userEloRepo.GetUserLeaderboard(r.Context(), userId, category, 100)
	}

	if err != nil {
		logger.Error("[Handler - Leaderboard] Couldn't fetch personal leaderboard. %v", err)
		return nil, "Error al carregar els resultats", 0
	}

	// Convert to display format
	entries := make([]LeaderboardEntry, len(apiEntries))
	for i, entry := range apiEntries {
		entries[i] = LeaderboardEntry{
			Rank:        entry.Rank,
			TorronId:    entry.TorronId,
			TorronName:  entry.TorronName,
			TorronImage: entry.TorronImage,
			Rating:      entry.Rating,
			VoteCount:   entry.VoteCount,
		}
	}

	return entries, "", 0
}

// fetchGlobalLeaderboard gets community-wide rankings
func (h *Handler) fetchGlobalLeaderboard(r *http.Request, category string) ([]LeaderboardEntry, string) {
	// Fetch global leaderboard via torron repository
	var torrons []*domain.Torro
	var err error

	if category == "global" {
		// Get all torrons sorted by rating
		torrons, err = h.torroRepo.List(r.Context())
	} else {
		// Get torrons for specific class sorted by rating
		torrons, err = h.torroRepo.ListByClass(r.Context(), category)
	}

	if err != nil {
		logger.Error("[Handler - Leaderboard] Couldn't fetch torrons. %v", err)
		return nil, "Error al carregar els resultats"
	}

	// Convert to display format
	entries := make([]LeaderboardEntry, 0, len(torrons))
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
			VoteCount:   0, // Global view doesn't show per-user vote counts
		})
		rank++

		// Limit to top 100
		if rank > 100 {
			break
		}
	}

	return entries, ""
}

// getClassName is a helper to get class name by ID
func (h *Handler) getClassName(classes []*domain.Class, classId string) string {
	for _, class := range classes {
		if class.Id == classId {
			return class.Name
		}
	}
	return "Desconegut"
}
