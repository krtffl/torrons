package http

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// FriendsContent holds data for the friends page template. View selects
// which fragment of friends.html gets rendered; only the fields relevant to
// that view are populated.
type FriendsContent struct {
	HX   bool
	View string // "index" | "created" | "leaderboard" | "not-member" | "invalid-invite"

	// "index" view: circles the current user belongs to
	Circles []*domain.FriendCircle

	// "created" view: the just-created circle's shareable invite link
	InviteURL string

	// "leaderboard" / "not-member" views
	CircleId         string
	SelectedCategory string
	Categories       []*domain.Class
	Entries          []LeaderboardEntry
	Error            string
}

// friendsIndex lists the circles the current user belongs to (owned or
// joined) and offers a "create a new circle" action. Not one of the three
// routes named in the spec, but a minimal, necessary entry point: without
// it there would be no in-app way to reach POST /friends/create.
func (h *Handler) friendsIndex(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Friends] Incoming index request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Friends] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	circles, err := h.friendCircleRepo.ListForUser(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - Friends] Couldn't list circles. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.renderFriends(w, FriendsContent{
		HX:      isHX(r),
		View:    "index",
		Circles: circles,
	})
}

// friendsCreate creates a new circle owned by the current user and shows
// its shareable invite link.
func (h *Handler) friendsCreate(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Friends] Incoming create request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Friends] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	circle, err := h.friendCircleRepo.Create(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - Friends] Couldn't create circle. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	inviteURL := fmt.Sprintf("%s/friends/join/%s", baseURL(r), circle.InviteCode)

	h.renderFriends(w, FriendsContent{
		HX:        isHX(r),
		View:      "created",
		CircleId:  circle.Id,
		InviteURL: inviteURL,
	})
}

// friendsJoin adds the current (anonymous, cookie-identified) user to the
// circle behind an invite code, if they're not already a member, then
// redirects to the circle's leaderboard. This is meant to be opened as a
// normal top-level navigation (a link shared in chat/social), not an HTMX
// fragment request, so it issues a real redirect rather than swapping HTML.
func (h *Handler) friendsJoin(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Friends] Incoming join request")

	inviteCode := chi.URLParam(r, "inviteCode")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Friends] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	circle, err := h.friendCircleRepo.GetByInviteCode(r.Context(), inviteCode)
	if err != nil {
		logger.Warn("[Handler - Friends] Unknown invite code %s. %v", inviteCode, err)
		h.renderFriends(w, FriendsContent{
			HX:   isHX(r),
			View: "invalid-invite",
		})
		return
	}

	if err := h.friendCircleRepo.AddMember(r.Context(), circle.Id, userId); err != nil {
		logger.Error("[Handler - Friends] Couldn't add member. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/friends/"+circle.Id, http.StatusFound)
}

// friendsLeaderboard renders the personalized ELO leaderboard scoped to a
// circle's members, defaulting to the Global class with a class switcher
// matching leaderboard.html's pattern.
func (h *Handler) friendsLeaderboard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Friends] Incoming leaderboard request")

	circleId := chi.URLParam(r, "circleId")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Friends] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	circle, err := h.friendCircleRepo.Get(r.Context(), circleId)
	if err != nil {
		logger.Warn("[Handler - Friends] Circle not found %s. %v", circleId, err)
		http.Error(w, "Not Found", http.StatusNotFound)
		return
	}

	// Only members can view the circle's leaderboard; membership is only
	// granted via the invite link (GET /friends/join/{inviteCode}), never by
	// visiting this URL directly.
	isMember, err := h.friendCircleRepo.IsMember(r.Context(), circle.Id, userId)
	if err != nil {
		logger.Error("[Handler - Friends] Couldn't check membership. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if !isMember {
		h.renderFriends(w, FriendsContent{
			HX:       isHX(r),
			View:     "not-member",
			CircleId: circle.Id,
		})
		return
	}

	category := r.URL.Query().Get("category")
	if category == "" {
		category = "global"
	}

	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - Friends] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var apiEntries []*domain.UserLeaderboardEntry
	if category == "global" {
		apiEntries, err = h.friendCircleRepo.GetCircleGlobalLeaderboard(r.Context(), circle.Id)
	} else {
		apiEntries, err = h.friendCircleRepo.GetCircleLeaderboard(r.Context(), circle.Id, category)
	}
	if err != nil {
		logger.Error("[Handler - Friends] Couldn't fetch circle leaderboard. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	entries := make([]LeaderboardEntry, len(apiEntries))
	for i, e := range apiEntries {
		entries[i] = LeaderboardEntry{
			Rank:        e.Rank,
			TorronId:    e.TorronId,
			TorronName:  e.TorronName,
			TorronImage: e.TorronImage,
			Rating:      e.Rating,
			VoteCount:   e.VoteCount,
		}
	}
	entries = calculateRatingPercentages(entries)

	errorMsg := ""
	if len(entries) == 0 {
		errorMsg = "Encara no hi ha prou vots dels membres d'aquest cercle per mostrar resultats en aquesta categoria"
	}

	h.renderFriends(w, FriendsContent{
		HX:               isHX(r),
		View:             "leaderboard",
		CircleId:         circle.Id,
		SelectedCategory: category,
		Categories:       classes,
		Entries:          entries,
		Error:            errorMsg,
	})
}

func (h *Handler) renderFriends(w http.ResponseWriter, content FriendsContent) {
	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "friends.html", content); err != nil {
		logger.Error("[Handler - Friends] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// baseURL derives the public origin (scheme + host) from the incoming
// request so invite links work the same in local dev and in production
// without hardcoding a domain.
func baseURL(r *http.Request) string {
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, r.Host)
}
