package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/krtffl/torro/internal/logger"
)

// CountdownResponse contains countdown information for the active campaign
type CountdownResponse struct {
	CampaignId       string `json:"campaign_id"`
	CampaignName     string `json:"campaign_name"`
	EndDate          string `json:"end_date"`
	TimeRemaining    int64  `json:"time_remaining_seconds"`
	IsActive         bool   `json:"is_active"`
	HasEnded         bool   `json:"has_ended"`
	DaysRemaining    int    `json:"days_remaining"`
	HoursRemaining   int    `json:"hours_remaining"`
	MinutesRemaining int    `json:"minutes_remaining"`
	SecondsRemaining int    `json:"seconds_remaining"`
}

// GlobalLeaderboardEntry represents a torron in the global leaderboard
type GlobalLeaderboardEntry struct {
	Rank       int     `json:"rank"`
	TorronId   string  `json:"torron_id"`
	TorronName string  `json:"torron_name"`
	Image      string  `json:"image"`
	Rating     float64 `json:"rating"`
	ClassName  string  `json:"class_name"`
}

// handleCountdown returns countdown information for the active campaign
func (h *Handler) handleCountdown(w http.ResponseWriter, r *http.Request) {
	// Get active campaign
	campaign, err := h.campaignRepo.GetActive(r.Context())
	if err != nil {
		// No active campaign
		render.Status(r, http.StatusOK)
		render.JSON(w, r, CountdownResponse{
			IsActive: false,
			HasEnded: true,
		})
		return
	}

	// Parse end date
	endTime, err := time.Parse(time.RFC3339, campaign.EndDate)
	if err != nil {
		logger.Error("[Countdown API] Invalid end date format: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Invalid campaign configuration"})
		return
	}

	now := time.Now().UTC()
	timeRemaining := endTime.Sub(now)

	// Check if campaign has ended
	hasEnded := timeRemaining <= 0
	if hasEnded {
		timeRemaining = 0
	}

	// Calculate days, hours, minutes, seconds
	days := int(timeRemaining.Hours() / 24)
	hours := int(timeRemaining.Hours()) % 24
	minutes := int(timeRemaining.Minutes()) % 60
	seconds := int(timeRemaining.Seconds()) % 60

	response := CountdownResponse{
		CampaignId:       campaign.Id,
		CampaignName:     campaign.Name,
		EndDate:          campaign.EndDate,
		TimeRemaining:    int64(timeRemaining.Seconds()),
		IsActive:         !hasEnded,
		HasEnded:         hasEnded,
		DaysRemaining:    days,
		HoursRemaining:   hours,
		MinutesRemaining: minutes,
		SecondsRemaining: seconds,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// handleCountdownWidget returns the countdown widget HTML for homepage
func (h *Handler) handleCountdownWidget(w http.ResponseWriter, r *http.Request) {
	// Get active campaign
	campaign, err := h.campaignRepo.GetActive(r.Context())

	content := struct {
		IsActive         bool
		HasEnded         bool
		DaysRemaining    int
		HoursRemaining   int
		MinutesRemaining int
		EndDateFormatted string
	}{
		IsActive: false,
		HasEnded: true,
	}

	if err == nil {
		// Parse end date
		endTime, err := time.Parse(time.RFC3339, campaign.EndDate)
		if err == nil {
			now := time.Now().UTC()
			timeRemaining := endTime.Sub(now)

			// Check if campaign has ended
			hasEnded := timeRemaining <= 0

			if !hasEnded {
				// Calculate days, hours, minutes
				days := int(timeRemaining.Hours() / 24)
				hours := int(timeRemaining.Hours()) % 24
				minutes := int(timeRemaining.Minutes()) % 60

				// Format end date in Catalan format
				endDateFormatted := endTime.Format("2 January 2006")

				content = struct {
					IsActive         bool
					HasEnded         bool
					DaysRemaining    int
					HoursRemaining   int
					MinutesRemaining int
					EndDateFormatted string
				}{
					IsActive:         true,
					HasEnded:         false,
					DaysRemaining:    days,
					HoursRemaining:   hours,
					MinutesRemaining: minutes,
					EndDateFormatted: endDateFormatted,
				}
			}
		}
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "countdown.html", content); err != nil {
		logger.Error("[Handler - CountdownWidget] Couldn't execute template. %v", err)
		w.Write([]byte("<div class='countdown-error'>Error carregant comptador</div>"))
		return
	}

	buf.WriteTo(w)
}

// handleGlobalLeaderboard returns the global leaderboard (all categories)
func (h *Handler) handleGlobalLeaderboard(w http.ResponseWriter, r *http.Request) {
	// Get top torrons by rating across all categories
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT
			t."Id",
			t."Name",
			t."Image",
			t."Rating",
			c."Name" as class_name,
			RANK() OVER (ORDER BY t."Rating" DESC) as rank
		 FROM "Torrons" t
		 INNER JOIN "Classes" c ON t."Class" = c."Id"
		 WHERE t."Discontinued" = false
		 ORDER BY t."Rating" DESC
		 LIMIT 100`,
	)
	if err != nil {
		logger.Error("[Global Leaderboard] Query error: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var entries []GlobalLeaderboardEntry
	for rows.Next() {
		entry := GlobalLeaderboardEntry{}
		err := rows.Scan(
			&entry.TorronId,
			&entry.TorronName,
			&entry.Image,
			&entry.Rating,
			&entry.ClassName,
			&entry.Rank,
		)
		if err != nil {
			logger.Error("[Global Leaderboard] Scan error: %v", err)
			continue
		}
		entries = append(entries, entry)
	}

	response := map[string]interface{}{
		"entries":       entries,
		"total_entries": len(entries),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// handleClassLeaderboard returns leaderboard for a specific class
func (h *Handler) handleClassLeaderboard(w http.ResponseWriter, r *http.Request) {
	classId := chi.URLParam(r, "classId")
	if classId == "" {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Class ID required"})
		return
	}

	// Get class info
	class, err := h.classRepo.Get(r.Context(), classId)
	if err != nil {
		logger.Error("[Class Leaderboard] Class not found: %v", err)
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "Class not found"})
		return
	}

	// Get top torrons by rating in this class
	rows, err := h.db.QueryContext(r.Context(),
		`SELECT
			t."Id",
			t."Name",
			t."Image",
			t."Rating",
			RANK() OVER (ORDER BY t."Rating" DESC) as rank
		 FROM "Torrons" t
		 WHERE t."Class" = $1
		   AND t."Discontinued" = false
		 ORDER BY t."Rating" DESC`,
		classId,
	)
	if err != nil {
		logger.Error("[Class Leaderboard] Query error: %v", err)
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Internal server error"})
		return
	}
	defer rows.Close()

	var entries []GlobalLeaderboardEntry
	for rows.Next() {
		entry := GlobalLeaderboardEntry{}
		err := rows.Scan(
			&entry.TorronId,
			&entry.TorronName,
			&entry.Image,
			&entry.Rating,
			&entry.Rank,
		)
		if err != nil {
			logger.Error("[Class Leaderboard] Scan error: %v", err)
			continue
		}
		entry.ClassName = class.Name
		entries = append(entries, entry)
	}

	response := map[string]interface{}{
		"class_id":      classId,
		"class_name":    class.Name,
		"entries":       entries,
		"total_entries": len(entries),
		"timestamp":     time.Now().UTC().Format(time.RFC3339),
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}

// handleCampaignInfo returns information about the active campaign
func (h *Handler) handleCampaignInfo(w http.ResponseWriter, r *http.Request) {
	campaign, err := h.campaignRepo.GetActive(r.Context())
	if err != nil {
		render.Status(r, http.StatusNotFound)
		render.JSON(w, r, map[string]string{"error": "No active campaign"})
		return
	}

	response := map[string]interface{}{
		"id":          campaign.Id,
		"name":        campaign.Name,
		"start_date":  campaign.StartDate,
		"end_date":    campaign.EndDate,
		"status":      campaign.Status,
		"year":        campaign.Year,
		"description": campaign.Description,
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)
}
