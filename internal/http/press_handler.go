package http

import (
	"math"
	"net/http"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// pressRiserWindowDays is the lookback window for the "biggest riser" stat.
const pressRiserWindowDays = 7

// pressClosestDuelMinVotes is the minimum total-vote threshold for a
// pairing to be eligible as the "closest duel" (avoids a 1-0 "duel" reading
// as maximally close).
const pressClosestDuelMinVotes = 10

// PressEmbedCategory is one option in the "embed the leaderboard" category
// picker on the press page.
type PressEmbedCategory struct {
	Id   string
	Name string
}

// PressContent holds data for the /premsa page template. Every "Has*" flag
// resolves the corresponding stat's legitimate empty state (e.g. no votes
// cast yet) so the template never has to deal with nil checks directly -
// same convention as TorroDetail in torro_handler.go.
type PressContent struct {
	HX bool

	HasMostVoted   bool
	MostVotedName  string
	MostVotedImage string
	MostVotedVotes int

	HasBiggestRiser bool
	RiserName       string
	RiserImage      string
	RiserPoints     int // net rating change over the window, rounded; can be negative

	HasClosestDuel  bool
	DuelAName       string
	DuelAImage      string
	DuelAPercentage int
	DuelBName       string
	DuelBImage      string
	DuelBPercentage int
	DuelTotalVotes  int

	HasChampion   bool
	ChampionName  string
	ChampionImage string

	EmbedCategories     []PressEmbedCategory
	EmbedDefaultClassId string
	EmbedBaseURL        string
}

// press renders the /premsa page: a small set of always-fresh, screenshot-
// friendly stats for journalists (most voted torró, biggest riser, closest
// duel, Gran Final result), plus a self-service snippet generator for
// embedding the live leaderboard widget (see embed_handler.go) on a third
// party's site.
func (h *Handler) press(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Press] Incoming request")

	ctx := r.Context()

	content := PressContent{
		HX:                  isHX(r),
		EmbedDefaultClassId: embedDefaultClassId,
		EmbedBaseURL:        baseURL(r),
	}

	mostVoted, err := h.pressStatsRepo.MostVotedTorro(ctx)
	if err != nil {
		logger.Error("[Handler - Press] Couldn't fetch most voted torró. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if mostVoted != nil {
		content.HasMostVoted = true
		content.MostVotedName = mostVoted.Name
		content.MostVotedImage = mostVoted.Image
		content.MostVotedVotes = int(math.Round(mostVoted.Value))
	}

	riser, err := h.pressStatsRepo.BiggestRiser(ctx, pressRiserWindowDays)
	if err != nil {
		logger.Error("[Handler - Press] Couldn't fetch biggest riser. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if riser != nil {
		content.HasBiggestRiser = true
		content.RiserName = riser.Name
		content.RiserImage = riser.Image
		content.RiserPoints = int(math.Round(riser.Value))
	}

	duel, err := h.pressStatsRepo.ClosestDuel(ctx, pressClosestDuelMinVotes)
	if err != nil {
		logger.Error("[Handler - Press] Couldn't fetch closest duel. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	if duel != nil {
		content.HasClosestDuel = true
		content.DuelAName = duel.TorroA.Name
		content.DuelAImage = duel.TorroA.Image
		content.DuelAPercentage = votePercentage(duel.TorroA.Value, duel.TotalVotes)
		content.DuelBName = duel.TorroB.Name
		content.DuelBImage = duel.TorroB.Image
		content.DuelBPercentage = votePercentage(duel.TorroB.Value, duel.TotalVotes)
		content.DuelTotalVotes = duel.TotalVotes
	}

	// The Gran Final is Phase 2's knockout bracket for the Global class,
	// separate from the Phase 1 ELO stats above. Following the existing
	// convention in bracket_handler.go's bracketOverview, any error here
	// (most commonly "no bracket created yet for this class") is treated
	// as the empty state rather than a hard failure.
	bracket, err := h.bracketRepo.GetLatestByClass(ctx, embedDefaultClassId)
	if err == nil && bracket != nil && bracket.Status == domain.BracketStatusCompleted && bracket.ChampionId != nil {
		champion, err := h.torroRepo.Get(ctx, *bracket.ChampionId)
		if err != nil {
			logger.Error("[Handler - Press] Couldn't fetch champion torró. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		content.HasChampion = true
		content.ChampionName = champion.Name
		content.ChampionImage = champion.Image
	}

	classes, err := h.classRepo.List(ctx)
	if err != nil {
		logger.Error("[Handler - Press] Couldn't list classes. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	content.EmbedCategories = make([]PressEmbedCategory, 0, len(classes))
	for _, c := range classes {
		content.EmbedCategories = append(content.EmbedCategories, PressEmbedCategory{
			Id:   c.Id,
			Name: c.Name,
		})
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "press.html", content); err != nil {
		logger.Error("[Handler - Press] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// votePercentage returns the integer percentage (0-100) that `value` votes
// represents out of `total`. Returns 0 if total is 0 to avoid a division by
// zero; ClosestDuel is filtered by a minimum vote threshold so this never
// actually happens in practice, but it keeps the helper safe regardless.
func votePercentage(value float64, total int) int {
	if total <= 0 {
		return 0
	}
	return int(math.Round(value / float64(total) * 100))
}
