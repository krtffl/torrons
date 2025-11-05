package http

import (
	"context"
	"net/http"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// CategoryProgress represents progress towards unlocking a category
type CategoryProgress struct {
	Id                 string
	Name               string
	Icon               string
	VoteCount          int
	MinVotes           int
	ProgressPercentage int
	VotesRemaining     int
	Unlocked           bool
}

// Achievement represents a user achievement/badge
type Achievement struct {
	Icon        string
	Name        string
	Description string
	Unlocked    bool
}

// StatsContent holds data for stats page template
type StatsContent struct {
	HX                  bool
	TotalVotes          int
	UnlockedCategories  int
	UserRank            string
	CategoryProgress    []CategoryProgress
	Achievements        []Achievement
}

// stats handles the user statistics page
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Stats] Incoming request")

	userId := getUserId(r)
	if userId == "" {
		logger.Error("[Handler - Stats] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get user data
	user, err := h.userRepo.Get(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - Stats] Couldn't get user. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get all classes
	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - Stats] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Category icons mapping
	categoryIcons := map[string]string{
		"1": "🏛️", // Clàssics
		"2": "✨", // Novetats
		"3": "🍫", // Xocolata
		"4": "👨‍🍳", // Albert Adrià
		"5": "🌍", // Global
	}

	// Build category progress
	var categoryProgress []CategoryProgress
	unlockedCount := 0

	for _, class := range classes {
		minVotes := getMinVotesForClass(class.Id)

		var voteCount int
		if class.Id == "5" { // Global uses total votes
			voteCount = user.VoteCount
		} else {
			voteCount, _ = h.userRepo.GetVoteCountForClass(r.Context(), userId, class.Id)
		}

		unlocked := voteCount >= minVotes
		if unlocked {
			unlockedCount++
		}

		votesRemaining := minVotes - voteCount
		if votesRemaining < 0 {
			votesRemaining = 0
		}

		progressPercentage := (voteCount * 100) / minVotes
		if progressPercentage > 100 {
			progressPercentage = 100
		}

		icon, ok := categoryIcons[class.Id]
		if !ok {
			icon = "📊"
		}

		categoryProgress = append(categoryProgress, CategoryProgress{
			Id:                 class.Id,
			Name:               class.Name,
			Icon:               icon,
			VoteCount:          voteCount,
			MinVotes:           minVotes,
			ProgressPercentage: progressPercentage,
			VotesRemaining:     votesRemaining,
			Unlocked:           unlocked,
		})
	}

	// Determine user rank based on total votes
	userRank := getUserRank(user.VoteCount)

	// Define achievements
	achievements := []Achievement{
		{
			Icon:        "🌟",
			Name:        "Primer vot",
			Description: "Has votat per primera vegada",
			Unlocked:    user.VoteCount >= 1,
		},
		{
			Icon:        "🔥",
			Name:        "Votant actiu",
			Description: "Has votat 25 vegades",
			Unlocked:    user.VoteCount >= 25,
		},
		{
			Icon:        "💯",
			Name:        "Centenari",
			Description: "Has votat 100 vegades",
			Unlocked:    user.VoteCount >= 100,
		},
		{
			Icon:        "🏆",
			Name:        "Expert en torrons",
			Description: "Has desbloquejat totes les categories",
			Unlocked:    unlockedCount == len(classes),
		},
		{
			Icon:        "🎯",
			Name:        "Completista",
			Description: "Has votat en totes les categories",
			Unlocked:    h.hasVotedAllCategories(r.Context(), userId, classes),
		},
		{
			Icon:        "⚡",
			Name:        "Súper votant",
			Description: "Has votat 200 vegades",
			Unlocked:    user.VoteCount >= 200,
		},
	}

	content := StatsContent{
		HX:                 isHX(r),
		TotalVotes:         user.VoteCount,
		UnlockedCategories: unlockedCount,
		UserRank:           userRank,
		CategoryProgress:   categoryProgress,
		Achievements:       achievements,
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "stats.html", content); err != nil {
		logger.Error("[Handler - Stats] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// getUserRank returns a rank label based on vote count
func getUserRank(voteCount int) string {
	switch {
	case voteCount >= 200:
		return "Mestre torronaire"
	case voteCount >= 100:
		return "Expert"
	case voteCount >= 50:
		return "Aficionat"
	case voteCount >= 25:
		return "Novell"
	case voteCount >= 10:
		return "Aprenent"
	default:
		return "Principiant"
	}
}

// hasVotedAllCategories checks if user has voted in all categories
func (h *Handler) hasVotedAllCategories(ctx context.Context, userId string, classes []*domain.Class) bool {
	for _, class := range classes {
		if class.Id == "5" {
			continue // Skip global for this check
		}

		voteCount, err := h.userRepo.GetVoteCountForClass(ctx, userId, class.Id)
		if err != nil || voteCount == 0 {
			return false
		}
	}
	return true
}
