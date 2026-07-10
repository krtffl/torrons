package http

import (
	"context"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/internal/sharecard"
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
	MostVotedId    string
	MostVotedName  string
	MostVotedImage string
	MostVotedVotes int

	HasBiggestRiser bool
	RiserId         string
	RiserName       string
	RiserImage      string
	RiserPoints     int // net rating change over the window, rounded; can be negative

	HasClosestDuel  bool
	DuelAId         string
	DuelAName       string
	DuelAImage      string
	DuelAPercentage int
	DuelBId         string
	DuelBName       string
	DuelBImage      string
	DuelBPercentage int
	DuelTotalVotes  int

	HasChampion   bool
	ChampionId    string
	ChampionName  string
	ChampionImage string

	EmbedCategories     []PressEmbedCategory
	EmbedDefaultClassId string
	EmbedBaseURL        string
}

// pressStatsBlock holds exactly the /premsa stats that come from the
// expensive, full-table-aggregating pressStatsRepo (plus the champion
// lookup): every "Has*" flag and stat field of PressContent EXCEPT the
// per-request ones (HX, EmbedCategories, EmbedDefaultClassId,
// EmbedBaseURL). These are a pure global aggregate - identical for every
// visitor - so they are memoized across requests (see pressCache) rather
// than recomputed on each hit.
type pressStatsBlock struct {
	HasMostVoted   bool
	MostVotedId    string
	MostVotedName  string
	MostVotedImage string
	MostVotedVotes int

	HasBiggestRiser bool
	RiserId         string
	RiserName       string
	RiserImage      string
	RiserPoints     int

	HasClosestDuel  bool
	DuelAId         string
	DuelAName       string
	DuelAImage      string
	DuelAPercentage int
	DuelBId         string
	DuelBName       string
	DuelBImage      string
	DuelBPercentage int
	DuelTotalVotes  int

	HasChampion   bool
	ChampionId    string
	ChampionName  string
	ChampionImage string
}

// pressStatsCacheTTL is how long a computed pressStatsBlock is served
// before it is refreshed. The /premsa stats run 3+ full-table aggregations
// over the ever-growing Results table plus a champion lookup (measured
// ~500ms solo, ~14s p95 under 50 concurrent), all producing a global
// aggregate that barely changes minute to minute - so a short in-process
// TTL collapses that to (at most) one refresh per minute.
const pressStatsCacheTTL = 60 * time.Second

// pressCache memoizes the global /premsa stat block across requests. The
// Handler is constructed once (see server.go), but these stats carry no
// per-Handler state - they're a pure global aggregate - so a package-level
// cache is both correct and the simplest option. `mu` guards the cached
// value; `refresh` serializes recomputation so a burst of requests arriving
// at expiry collapses into a single DB refresh instead of a stampede.
var pressCache struct {
	mu      sync.RWMutex
	block   pressStatsBlock
	expiry  time.Time
	hasData bool

	refresh sync.Mutex
}

// press renders the /premsa page: a small set of screenshot-friendly stats
// for journalists (most voted torró, biggest riser, closest duel, Gran
// Final result), plus a self-service snippet generator for embedding the
// live leaderboard widget (see embed_handler.go) on a third party's site.
// The stats are served from a short-TTL in-process cache (see pressStats);
// the embed-picker fields are always built fresh per request.
func (h *Handler) press(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Press] Incoming request")

	ctx := r.Context()

	stats, err := h.pressStats(ctx)
	if err != nil {
		logger.Error("[Handler - Press] Couldn't fetch press stats. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	content := PressContent{
		HX:                  isHX(r),
		EmbedDefaultClassId: embedDefaultClassId,
		EmbedBaseURL:        baseURL(r),

		HasMostVoted:   stats.HasMostVoted,
		MostVotedId:    stats.MostVotedId,
		MostVotedName:  stats.MostVotedName,
		MostVotedImage: stats.MostVotedImage,
		MostVotedVotes: stats.MostVotedVotes,

		HasBiggestRiser: stats.HasBiggestRiser,
		RiserId:         stats.RiserId,
		RiserName:       stats.RiserName,
		RiserImage:      stats.RiserImage,
		RiserPoints:     stats.RiserPoints,

		HasClosestDuel:  stats.HasClosestDuel,
		DuelAId:         stats.DuelAId,
		DuelAName:       stats.DuelAName,
		DuelAImage:      stats.DuelAImage,
		DuelAPercentage: stats.DuelAPercentage,
		DuelBId:         stats.DuelBId,
		DuelBName:       stats.DuelBName,
		DuelBImage:      stats.DuelBImage,
		DuelBPercentage: stats.DuelBPercentage,
		DuelTotalVotes:  stats.DuelTotalVotes,

		HasChampion:   stats.HasChampion,
		ChampionId:    stats.ChampionId,
		ChampionName:  stats.ChampionName,
		ChampionImage: stats.ChampionImage,
	}

	// EmbedCategories depends only on the fixed 5-row class catalog, but is
	// rebuilt per request (kept out of the cached stat block) so the embed
	// picker always reflects the live catalog.
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

// pressStats returns the cached global /premsa stat block, refreshing it
// when the TTL has expired. On a refresh error it keeps serving the last
// good cached value if one exists (so a transient DB hiccup doesn't take
// the page down), and only surfaces the error when there is nothing cached
// to fall back on.
func (h *Handler) pressStats(ctx context.Context) (pressStatsBlock, error) {
	// Fast path: a fresh cached value serves the vast majority of requests
	// under a single read lock.
	pressCache.mu.RLock()
	if pressCache.hasData && time.Now().Before(pressCache.expiry) {
		block := pressCache.block
		pressCache.mu.RUnlock()
		return block, nil
	}
	pressCache.mu.RUnlock()

	// Stale or cold: serialize the refresh so a burst arriving at expiry
	// triggers a single recomputation, not one expensive query per request.
	pressCache.refresh.Lock()
	defer pressCache.refresh.Unlock()

	// Re-check: another goroutine may have refreshed while we waited on the
	// refresh lock.
	pressCache.mu.RLock()
	if pressCache.hasData && time.Now().Before(pressCache.expiry) {
		block := pressCache.block
		pressCache.mu.RUnlock()
		return block, nil
	}
	pressCache.mu.RUnlock()

	block, err := h.computePressStats(ctx)
	if err != nil {
		// Fall back to the last good value on a transient failure.
		pressCache.mu.RLock()
		defer pressCache.mu.RUnlock()
		if pressCache.hasData {
			logger.Warn("[Handler - Press] Stats refresh failed; serving last cached value. %v", err)
			return pressCache.block, nil
		}
		return pressStatsBlock{}, err
	}

	pressCache.mu.Lock()
	pressCache.block = block
	pressCache.expiry = time.Now().Add(pressStatsCacheTTL)
	pressCache.hasData = true
	pressCache.mu.Unlock()

	return block, nil
}

// computePressStats runs the actual (expensive) aggregate queries that back
// the /premsa stats and assembles them into a pressStatsBlock. Any repo
// error is returned to the caller (pressStats), which decides whether to
// surface it or fall back to a cached value. The empty/no-votes states are
// preserved exactly as before: a nil result from a repo leaves the
// corresponding "Has*" flag false.
func (h *Handler) computePressStats(ctx context.Context) (pressStatsBlock, error) {
	var block pressStatsBlock

	mostVoted, err := h.pressStatsRepo.MostVotedTorro(ctx)
	if err != nil {
		return pressStatsBlock{}, err
	}
	if mostVoted != nil {
		block.HasMostVoted = true
		block.MostVotedId = mostVoted.TorroId
		block.MostVotedName = mostVoted.Name
		block.MostVotedImage = mostVoted.Image
		block.MostVotedVotes = int(math.Round(mostVoted.Value))
	}

	riser, err := h.pressStatsRepo.BiggestRiser(ctx, pressRiserWindowDays)
	if err != nil {
		return pressStatsBlock{}, err
	}
	if riser != nil {
		block.HasBiggestRiser = true
		block.RiserId = riser.TorroId
		block.RiserName = riser.Name
		block.RiserImage = riser.Image
		block.RiserPoints = int(math.Round(riser.Value))
	}

	duel, err := h.pressStatsRepo.ClosestDuel(ctx, pressClosestDuelMinVotes)
	if err != nil {
		return pressStatsBlock{}, err
	}
	if duel != nil {
		block.HasClosestDuel = true
		block.DuelAId = duel.TorroA.TorroId
		block.DuelAName = duel.TorroA.Name
		block.DuelAImage = duel.TorroA.Image
		block.DuelAPercentage = votePercentage(duel.TorroA.Value, duel.TotalVotes)
		block.DuelBId = duel.TorroB.TorroId
		block.DuelBName = duel.TorroB.Name
		block.DuelBImage = duel.TorroB.Image
		block.DuelBPercentage = votePercentage(duel.TorroB.Value, duel.TotalVotes)
		block.DuelTotalVotes = duel.TotalVotes
	}

	// The Gran Final is Phase 2's knockout bracket for the Global class,
	// separate from the Phase 1 ELO stats above. pressGlobalChampion
	// follows the existing convention in bracket_handler.go's
	// bracketOverview: any error resolving the latest bracket (most
	// commonly "no bracket created yet for this class") is treated as the
	// empty state rather than a hard failure.
	champion, err := h.pressGlobalChampion(ctx)
	if err != nil {
		return pressStatsBlock{}, err
	}
	if champion != nil {
		block.HasChampion = true
		block.ChampionId = champion.Id
		block.ChampionName = champion.Name
		block.ChampionImage = champion.Image
	}

	return block, nil
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

// pressGlobalChampion resolves the Global bracket's champion torró, if
// the Gran Final has been decided. Returns (nil, nil) - not an error -
// when there's no bracket yet, it's still in progress, or GetLatestByClass
// itself errors (following bracketOverview's precedent that "no bracket"
// is a legitimate empty state, not a failure). Shared by press (the page)
// and pressKitCard (the PNG one-pager) so the two surfaces can never
// disagree about who the champion is.
func (h *Handler) pressGlobalChampion(ctx context.Context) (*domain.Torro, error) {
	bracket, err := h.bracketRepo.GetLatestByClass(ctx, embedDefaultClassId)
	if err != nil || bracket == nil || bracket.Status != domain.BracketStatusCompleted || bracket.ChampionId == nil {
		return nil, nil
	}

	champion, err := h.torroRepo.Get(ctx, *bracket.ChampionId)
	if err != nil {
		return nil, err
	}

	return champion, nil
}

// pressKitCard renders and serves the aggregate press-kit PNG one-pager:
// the Gran Final's champion torró and its total winning vote count, or an
// empty state if the bracket hasn't produced a champion yet (see
// sharecard.RenderPressKit). Unlike wrappedCard, this card is a global
// aggregate - the exact same PNG for every viewer, no user ID involved -
// so it uses a short public cache instead of the per-user "private,
// no-store" precedent, mirroring GET /embed/leaderboard.
func (h *Handler) pressKitCard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - PressKitCard] Incoming request")

	ctx := r.Context()

	champion, err := h.pressGlobalChampion(ctx)
	if err != nil {
		logger.Error("[Handler - PressKitCard] Couldn't fetch champion torró. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var data sharecard.PressKitData
	if champion != nil {
		votes, err := h.pressStatsRepo.VotesForTorro(ctx, champion.Id)
		if err != nil {
			logger.Error("[Handler - PressKitCard] Couldn't fetch champion vote count. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		data = sharecard.PressKitData{
			HasChampion:   true,
			ChampionName:  champion.Name,
			ChampionVotes: votes,
		}
	}

	// Cap concurrent renders (each allocates a large RGBA and pegs a core):
	// shed load with 503 rather than piling up when the cap is saturated.
	if !sharecard.TryAcquireRenderSlot(r.Context()) {
		logger.Warn("[Handler - PressKitCard] Render slots saturated; shedding request.")
		w.Header().Set("Retry-After", "1")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer sharecard.ReleaseRenderSlot()

	png, err := sharecard.RenderPressKit(data)
	if err != nil {
		logger.Error("[Handler - PressKitCard] Couldn't render card. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=300")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		logger.Error("[Handler - PressKitCard] Couldn't write response. %v", err)
	}
}
