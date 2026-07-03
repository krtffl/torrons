package http

import (
	"context"
	"net/http"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/internal/sharecard"
)

// WrappedContent holds data for the /wrapped page template. It mirrors
// sharecard.WrappedData field-for-field (plus HX) on purpose: both the
// page and the PNG card are built from the exact same wrappedCardData
// call below, so they can never show different numbers.
type WrappedContent struct {
	HX bool

	HasEnoughVotes bool
	VotesRemaining int

	TotalVotes int

	HasContestedDuel    bool
	ContestedTorroAName string
	ContestedTorroBName string
	ContestedPercentA   int
	ContestedPercentB   int

	HasUnpopularPick  bool
	UnpopularPickName string
	UnpopularPercent  int

	HasBracketVotes       bool
	BracketRoundsVoted    int
	BracketMatchesDecided int
	BracketPicksCorrect   int
	HasChampion           bool
	ChampionName          string
	MatchedChampion       bool
}

// wrapped handles GET /wrapped: the personal "Torrorèndum Wrapped"
// campaign recap page. Gated behind the same minimum-vote threshold as
// the Global leaderboard (getMinVotesForClass); below it, the page
// renders an honest "not unlocked yet" state instead of fabricating
// numbers from too little data.
func (h *Handler) wrapped(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Wrapped] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Wrapped] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	data, err := h.wrappedCardData(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - Wrapped] Couldn't fetch wrapped data. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	content := wrappedContentFromCardData(data)
	content.HX = isHX(r)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "wrapped.html", content); err != nil {
		logger.Error("[Handler - Wrapped] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// wrappedCard renders and serves the personal Wrapped 1080x1920 PNG
// "story" card, using the exact same data as the wrapped page above.
func (h *Handler) wrappedCard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - WrappedCard] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - WrappedCard] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	data, err := h.wrappedCardData(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - WrappedCard] Couldn't fetch wrapped data. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	png, err := sharecard.RenderWrapped(data)
	if err != nil {
		logger.Error("[Handler - WrappedCard] Couldn't render card. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Personalized per-user content: never cache/share across users, same
	// precedent as shareCard in sharecard_handler.go.
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		logger.Error("[Handler - WrappedCard] Couldn't write response. %v", err)
	}
}

// wrappedCardData assembles the single canonical sharecard.WrappedData
// value for one user, shared by both wrapped (HTML) and wrappedCard
// (PNG) so their data-fetching logic never has to be duplicated or drift
// apart.
func (h *Handler) wrappedCardData(ctx context.Context, userId string) (sharecard.WrappedData, error) {
	user, err := h.userRepo.Get(ctx, userId)
	if err != nil {
		return sharecard.WrappedData{}, err
	}

	minVotes := getMinVotesForClass(embedDefaultClassId) // "5" - Global
	if user.VoteCount < minVotes {
		return sharecard.WrappedData{
			HasEnoughVotes: false,
			VotesRemaining: minVotes - user.VoteCount,
			TotalVotes:     user.VoteCount,
		}, nil
	}

	data := sharecard.WrappedData{
		HasEnoughVotes: true,
		TotalVotes:     user.VoteCount,
	}

	duelStats, err := h.wrappedStatsRepo.DuelStats(ctx, userId, pressClosestDuelMinVotes)
	if err != nil {
		return sharecard.WrappedData{}, err
	}
	if duelStats.HasContestedDuel {
		data.HasContestedDuel = true
		data.ContestedTorroAName = duelStats.ContestedDuel.TorroAName
		data.ContestedTorroBName = duelStats.ContestedDuel.TorroBName
		data.ContestedPercentA = votePercentage(float64(duelStats.ContestedDuel.VotesA), duelStats.ContestedDuel.TotalVotes)
		data.ContestedPercentB = votePercentage(float64(duelStats.ContestedDuel.VotesB), duelStats.ContestedDuel.TotalVotes)
	}
	if duelStats.HasUnpopularPick {
		data.HasUnpopularPick = true
		data.UnpopularPickName = duelStats.UnpopularPick.UserPickName
		data.UnpopularPercent = userPickSharePercentage(duelStats.UnpopularPick)
	}

	// Following press_handler.go's bracketOverview precedent: any error
	// getting the latest bracket (most commonly "no bracket created yet")
	// is treated as "this user has no bracket participation to show", not
	// a hard failure.
	bracket, err := h.bracketRepo.GetLatestByClass(ctx, embedDefaultClassId)
	if err == nil && bracket != nil {
		bracketPath, err := h.wrappedStatsRepo.BracketPath(ctx, userId, bracket.Id)
		if err != nil {
			return sharecard.WrappedData{}, err
		}

		if bracketPath.HasVoted {
			data.HasBracketVotes = true
			data.BracketRoundsVoted = bracketPath.RoundsVoted
			data.BracketMatchesDecided = bracketPath.MatchesDecided
			data.BracketPicksCorrect = bracketPath.PicksCorrect
		}
		if bracketPath.HasChampion {
			data.HasChampion = true
			data.ChampionName = bracketPath.ChampionName
			data.MatchedChampion = bracketPath.MatchedChampion
		}
	}

	return data, nil
}

// userPickSharePercentage returns the % of the crowd that agreed with the
// user's own pick in a duel stat, i.e. whichever of VotesA/VotesB
// corresponds to UserPickId. Compares by ID (not name) so this stays
// correct even if the catalog ever has two distinct torrons sharing a
// display name.
func userPickSharePercentage(pick domain.UserDuelStat) int {
	userVotes := pick.VotesB
	if pick.UserPickId == pick.TorroAId {
		userVotes = pick.VotesA
	}
	return votePercentage(float64(userVotes), pick.TotalVotes)
}

// wrappedContentFromCardData copies sharecard.WrappedData's fields into
// the template-facing WrappedContent shape. Kept as a pure field copy (no
// extra derivation) so wrapped.html and the PNG card are guaranteed to
// show the exact same numbers.
func wrappedContentFromCardData(data sharecard.WrappedData) WrappedContent {
	return WrappedContent{
		HasEnoughVotes:        data.HasEnoughVotes,
		VotesRemaining:        data.VotesRemaining,
		TotalVotes:            data.TotalVotes,
		HasContestedDuel:      data.HasContestedDuel,
		ContestedTorroAName:   data.ContestedTorroAName,
		ContestedTorroBName:   data.ContestedTorroBName,
		ContestedPercentA:     data.ContestedPercentA,
		ContestedPercentB:     data.ContestedPercentB,
		HasUnpopularPick:      data.HasUnpopularPick,
		UnpopularPickName:     data.UnpopularPickName,
		UnpopularPercent:      data.UnpopularPercent,
		HasBracketVotes:       data.HasBracketVotes,
		BracketRoundsVoted:    data.BracketRoundsVoted,
		BracketMatchesDecided: data.BracketMatchesDecided,
		BracketPicksCorrect:   data.BracketPicksCorrect,
		HasChampion:           data.HasChampion,
		ChampionName:          data.ChampionName,
		MatchedChampion:       data.MatchedChampion,
	}
}
