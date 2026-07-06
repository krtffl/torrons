package http

import (
	"context"
	"html/template"
	"net/http"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// CategoryProgress represents progress towards unlocking a category
type CategoryProgress struct {
	Id                 string
	Name               string
	Tag                string
	Icon               template.HTML
	VoteCount          int
	MinVotes           int
	ProgressPercentage int
	VotesRemaining     int
	Unlocked           bool
}

// Achievement represents a user achievement/badge
type Achievement struct {
	Icon        template.HTML
	Name        string
	Description string
	Unlocked    bool
}

// Category and achievement icons below are hardcoded server-side SVG
// markup (not user input), so assigning them straight to template.HTML
// fields is safe — html/template would otherwise escape raw "<svg>..."
// strings into literal text instead of rendering them as markup.
// All icons share the same viewBox/stroke-based line-art convention for
// visual consistency; color is inherited from currentColor, which the
// surrounding .category-progress-icon / .achievement-icon CSS rules set.
const (
	iconCategoryClassics = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M3 8 12 3l9 5"/><line x1="4" y1="8" x2="20" y2="8"/><line x1="6" y1="8" x2="6" y2="18"/><line x1="12" y1="8" x2="12" y2="18"/><line x1="18" y1="8" x2="18" y2="18"/><line x1="3" y1="19" x2="21" y2="19"/><line x1="3" y1="21" x2="21" y2="21"/></svg>`)

	iconCategoryNovetats = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 3 13.8 9.8 21 12 13.8 14.2 12 21 10.2 14.2 3 12 10.2 9.8Z"/></svg>`)

	iconCategoryXocolata = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><rect x="3" y="3" width="18" height="18" rx="2"/><line x1="3" y1="12" x2="21" y2="12"/><line x1="12" y1="3" x2="12" y2="21"/></svg>`)

	iconCategoryAdriaNatura = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M7 10c-2.2 0-4-1.8-4-3.6C3 4.6 4.6 3 6.5 3c.9 0 1.7.35 2.3.9C9.6 2.7 10.7 2 12 2s2.4.7 3.2 1.9c.6-.55 1.4-.9 2.3-.9C19.4 3 21 4.6 21 6.4c0 1.8-1.8 3.6-4 3.6"/><path d="M7 10v3a5 5 0 0 0 5 5 5 5 0 0 0 5-5v-3"/><line x1="6" y1="19" x2="18" y2="19"/><line x1="6" y1="21" x2="18" y2="21"/></svg>`)

	iconCategoryGlobal = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9"/><ellipse cx="12" cy="12" rx="4" ry="9"/><line x1="3" y1="12" x2="21" y2="12"/><path d="M4.5 7.5c1.8 1.2 4.6 2 7.5 2s5.7-.8 7.5-2"/><path d="M4.5 16.5c1.8-1.2 4.6-2 7.5-2s5.7.8 7.5 2"/></svg>`)

	iconCategoryFallback = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><line x1="4" y1="20" x2="20" y2="20"/><rect x="6" y="13" width="3" height="7"/><rect x="11" y="9" width="3" height="11"/><rect x="16" y="5" width="3" height="15"/></svg>`)

	iconAchievementFirstVote = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 3 14.5 9 21 9.7 16 14 17.5 20.5 12 17 6.5 20.5 8 14 3 9.7 9.5 9Z"/></svg>`)

	iconAchievementActiveVoter = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M12 2c2 3-1 4-1 7 0 2 1.5 3 3 3s3-1 3-3c0-1-.4-1.8-1-2.4.2 1-.3 1.9-1.2 1.9-.9 0-1.4-.8-1.1-1.7C15.3 5 13.5 4 12 2Z"/><path d="M9 12c-.6.9-1 2-1 3a4 4 0 0 0 8 0c0-1-.4-2.1-1-3-.2 2-1.7 3.4-3.5 3.4S9.2 14 9 12Z"/></svg>`)

	iconAchievementCentenari = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="9" r="6"/><path d="M9 14.5 7 21l5-2.5L17 21l-2-6.5"/></svg>`)

	iconAchievementExpert = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M7 4h10v4a5 5 0 0 1-10 0V4Z"/><path d="M7 5H4a3 3 0 0 0 3 4"/><path d="M17 5h3a3 3 0 0 1-3 4"/><line x1="12" y1="13" x2="12" y2="17"/><line x1="9" y1="20" x2="15" y2="20"/><line x1="10" y1="17" x2="14" y2="17"/></svg>`)

	iconAchievementCompletista = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><circle cx="12" cy="12" r="9"/><path d="M8 12.5 10.5 15 16 9"/></svg>`)

	iconAchievementSuperVotant = template.HTML(`<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M13 2 4 14h6l-1 8 9-12h-6l1-8Z"/></svg>`)
)

// StatsContent holds data for stats page template
type StatsContent struct {
	HX                 bool
	TotalVotes         int
	UnlockedCategories int
	UserRank           string
	CategoryProgress   []CategoryProgress
	Achievements       []Achievement
	CurrentStreak      int
	LongestStreak      int
}

// stats handles the user statistics page
func (h *Handler) stats(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Stats] Incoming request")

	userId := GetUserIDFromContext(r.Context())
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
	categoryIcons := map[string]template.HTML{
		"1": iconCategoryClassics,    // Clàssics
		"2": iconCategoryNovetats,    // Novetats
		"3": iconCategoryXocolata,    // Xocolata
		"4": iconCategoryAdriaNatura, // Adrià Natura
		"5": iconCategoryGlobal,      // Global
	}

	// Category tag copy — ported verbatim from classes.html's .arena-tag
	// mapping (same per-id copy, same wording) so the two pages agree.
	categoryTags := map[string]string{
		"1": "L'ORIGINAL",
		"2": "EN TENDÈNCIA",
		"3": "PER ALS GOLAFRES",
		"4": "EDICIÓ LIMITADA",
		"5": "EL REPTE DEFINITIU",
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
			icon = iconCategoryFallback
		}

		categoryProgress = append(categoryProgress, CategoryProgress{
			Id:                 class.Id,
			Name:               class.Name,
			Tag:                categoryTags[class.Id],
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
			Icon:        iconAchievementFirstVote,
			Name:        "Primer vot",
			Description: "Has votat per primera vegada",
			Unlocked:    user.VoteCount >= 1,
		},
		{
			Icon:        iconAchievementActiveVoter,
			Name:        "Votant actiu",
			Description: "Has votat 25 vegades",
			Unlocked:    user.VoteCount >= 25,
		},
		{
			Icon:        iconAchievementCentenari,
			Name:        "Centenari",
			Description: "Has votat 100 vegades",
			Unlocked:    user.VoteCount >= 100,
		},
		{
			Icon:        iconAchievementExpert,
			Name:        "Expert en torrons",
			Description: "Has desbloquejat totes les categories",
			Unlocked:    unlockedCount == len(classes),
		},
		{
			Icon:        iconAchievementCompletista,
			Name:        "Completista",
			Description: "Has votat en totes les categories",
			Unlocked:    h.hasVotedAllCategories(r.Context(), userId, classes),
		},
		{
			Icon:        iconAchievementSuperVotant,
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
		CurrentStreak:      user.CurrentStreak,
		LongestStreak:      user.LongestStreak,
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
