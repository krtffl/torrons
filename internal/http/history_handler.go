package http

import (
	"database/sql"
	"net/http"
	"strconv"
	"time"

	"github.com/krtffl/torro/internal/logger"
)

// VoteHistory represents a single vote in the user's history
type VoteHistory struct {
	Torron1Name  string
	Torron1Image string
	Torron2Name  string
	Torron2Image string
	WinnerId     string
	IsWinner1    bool
	IsWinner2    bool
	CategoryName string
	CategoryIcon string
	Timestamp    time.Time
	TimeAgo      string
}

// HistoryContent holds data for history page template
type HistoryContent struct {
	HX             bool
	Votes          []VoteHistory
	FilterCategory string
	Categories     []CategoryWithIcon
	HasMore        bool
	NextOffset     int
}

// CategoryWithIcon includes icon for display
type CategoryWithIcon struct {
	Id   string
	Name string
	Icon string
}

// history handles the voting history page
func (h *Handler) history(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - History] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - History] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	// Get filter parameters
	filterCategory := r.URL.Query().Get("category")
	if filterCategory == "" {
		filterCategory = "all"
	}

	offsetStr := r.URL.Query().Get("offset")
	offset := 0
	if offsetStr != "" {
		offset, _ = strconv.Atoi(offsetStr)
	}

	limit := 20 // Show 20 votes per page

	// Build query
	var query string
	var args []interface{}

	baseQuery := `
		SELECT
			t1."Name" as torron1_name,
			t1."Image" as torron1_image,
			t2."Name" as torron2_name,
			t2."Image" as torron2_image,
			r."WinnerId",
			r."Timestamp",
			c."Name" as category_name,
			c."Id" as category_id
		FROM "Results" r
		INNER JOIN "Pairings" p ON r."PairingId" = p."Id"
		INNER JOIN "Torrons" t1 ON p."Torro1" = t1."Id"
		INNER JOIN "Torrons" t2 ON p."Torro2" = t2."Id"
		INNER JOIN "Classes" c ON p."Class" = c."Id"
		WHERE r."UserId" = $1
	`

	if filterCategory != "all" {
		query = baseQuery + " AND p.\"Class\" = $2 ORDER BY r.\"Timestamp\" DESC LIMIT $3 OFFSET $4"
		args = []interface{}{userId, filterCategory, limit + 1, offset} // +1 to check if there are more
	} else {
		query = baseQuery + " ORDER BY r.\"Timestamp\" DESC LIMIT $2 OFFSET $3"
		args = []interface{}{userId, limit + 1, offset}
	}

	rows, err := h.db.QueryContext(r.Context(), query, args...)
	if err != nil {
		logger.Error("[Handler - History] Query error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var votes []VoteHistory
	for rows.Next() {
		var vote VoteHistory
		var timestamp sql.NullTime
		var categoryId string

		err := rows.Scan(
			&vote.Torron1Name,
			&vote.Torron1Image,
			&vote.Torron2Name,
			&vote.Torron2Image,
			&vote.WinnerId,
			&timestamp,
			&vote.CategoryName,
			&categoryId,
		)
		if err != nil {
			logger.Error("[Handler - History] Scan error: %v", err)
			continue
		}

		if timestamp.Valid {
			vote.Timestamp = timestamp.Time
			vote.TimeAgo = getTimeAgo(timestamp.Time)
		}

		// Set winner flags
		vote.IsWinner1 = vote.WinnerId != "" && vote.Torron1Name != "" // WinnerId matches first torron
		vote.IsWinner2 = !vote.IsWinner1

		// Get category icon
		vote.CategoryIcon = getCategoryIcon(categoryId)

		votes = append(votes, vote)

		// Stop if we've reached the limit
		if len(votes) > limit {
			break
		}
	}

	// Check if there are more results
	hasMore := len(votes) > limit
	if hasMore {
		votes = votes[:limit] // Trim to actual limit
	}

	// Get all categories for filter
	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - History] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var categoriesWithIcons []CategoryWithIcon
	for _, class := range classes {
		categoriesWithIcons = append(categoriesWithIcons, CategoryWithIcon{
			Id:   class.Id,
			Name: class.Name,
			Icon: getCategoryIcon(class.Id),
		})
	}

	content := HistoryContent{
		HX:             isHX(r),
		Votes:          votes,
		FilterCategory: filterCategory,
		Categories:     categoriesWithIcons,
		HasMore:        hasMore,
		NextOffset:     offset + limit,
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "history.html", content); err != nil {
		logger.Error("[Handler - History] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// getTimeAgo returns a human-friendly time ago string
func getTimeAgo(t time.Time) string {
	duration := time.Since(t)

	if duration < time.Minute {
		return "Ara mateix"
	} else if duration < time.Hour {
		minutes := int(duration.Minutes())
		if minutes == 1 {
			return "Fa 1 minut"
		}
		return "Fa " + strconv.Itoa(minutes) + " minuts"
	} else if duration < 24*time.Hour {
		hours := int(duration.Hours())
		if hours == 1 {
			return "Fa 1 hora"
		}
		return "Fa " + strconv.Itoa(hours) + " hores"
	} else {
		days := int(duration.Hours() / 24)
		if days == 1 {
			return "Fa 1 dia"
		}
		return "Fa " + strconv.Itoa(days) + " dies"
	}
}

// getCategoryIcon returns emoji for category
func getCategoryIcon(categoryId string) string {
	icons := map[string]string{
		"1": "🏛️",
		"2": "✨",
		"3": "🍫",
		"4": "👨‍🍳",
		"5": "🌍",
	}
	icon, ok := icons[categoryId]
	if !ok {
		return "📊"
	}
	return icon
}
