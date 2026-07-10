package http

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/oxtoacart/bpool"

	torrons "github.com/krtffl/torro"
	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// K is the K-factor for ELO rating calculations
// Value of 42 provides:
// - Fast convergence for new items (larger rating changes)
// - Balanced sensitivity for established items
// - Standard K-factor is 32 for masters, 40 for beginners
// - 42 chosen for this system's moderate volatility needs
const K = 42

type Content struct {
	Torrons  []*domain.Torro
	Pairing  *domain.Pairing
	Pairings []*domain.Pairing
	Classes  []*domain.Class
	HX       bool

	// CurrentStreak/StreakAtRisk feed the small streak-indicator pill shown
	// near the top of the vote screen (design prompt 13). Both are zero
	// values for anonymous sessions or if the user lookup fails, which just
	// hides the pill -- it's decorative and never blocks voting.
	CurrentStreak int
	StreakAtRisk  bool
}

type Handler struct {
	db               *sql.DB
	template         *template.Template
	bpool            *bpool.BufferPool
	pairingRepo      domain.PairingRepo
	torroRepo        domain.TorroRepo
	classRepo        domain.ClassRepo
	resultRepo       domain.ResultRepo
	userRepo         domain.UserRepo
	userEloRepo      domain.UserEloSnapshotRepo
	campaignRepo     domain.CampaignRepo
	bracketRepo      domain.BracketRepo
	adventVoteRepo   domain.AdventVoteRepo
	friendCircleRepo domain.FriendCircleRepo
	pressStatsRepo   domain.PressStatsRepo
	wrappedStatsRepo domain.WrappedStatsRepo
	personaRepo      domain.PersonaRepo
	adminToken       string
}

func NewHandler(
	db *sql.DB,
	bpool *bpool.BufferPool,
	pairingRep domain.PairingRepo,
	torroRepo domain.TorroRepo,
	classRepo domain.ClassRepo,
	resultRepo domain.ResultRepo,
	userRepo domain.UserRepo,
	userEloRepo domain.UserEloSnapshotRepo,
	campaignRepo domain.CampaignRepo,
	bracketRepo domain.BracketRepo,
	adventVoteRepo domain.AdventVoteRepo,
	friendCircleRepo domain.FriendCircleRepo,
	pressStatsRepo domain.PressStatsRepo,
	wrappedStatsRepo domain.WrappedStatsRepo,
	personaRepo domain.PersonaRepo,
	adminToken string,
) *Handler {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		logger.Fatal("[Handler] - Failed to parse templates. %v", err)
	}

	return &Handler{
		db:               db,
		template:         tmpls,
		bpool:            bpool,
		pairingRepo:      pairingRep,
		torroRepo:        torroRepo,
		classRepo:        classRepo,
		resultRepo:       resultRepo,
		userRepo:         userRepo,
		userEloRepo:      userEloRepo,
		campaignRepo:     campaignRepo,
		bracketRepo:      bracketRepo,
		adventVoteRepo:   adventVoteRepo,
		friendCircleRepo: friendCircleRepo,
		pressStatsRepo:   pressStatsRepo,
		wrappedStatsRepo: wrappedStatsRepo,
		personaRepo:      personaRepo,
		adminToken:       adminToken,
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Index] Incoming request")

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	// Purely static (no DB dependency) - safe to cache. Same 1-hour TTL used
	// for /public/* CSS/JS: short enough that a deploy propagates quickly,
	// long enough to actually get cached. Vary on HX-Request because the
	// template renders a full page shell vs. an htmx partial depending on it -
	// without Vary a shared cache could serve the wrong variant to a client
	// with a different HX-Request header than whoever populated the cache.
	setStaticPageCacheHeaders(w)

	if err := h.template.ExecuteTemplate(buf, "index.html", Content{
		HX: isHX(r),
	}); err != nil {
		logger.Error("[Handler - Index] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - Index] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

func (h *Handler) classes(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Classes] Incoming request")

	classes, err := h.classRepo.List(r.Context())
	if err != nil {
		logger.Error("[Handler - Classes] Couldn't list classes. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "classes.html", Content{
		Classes: classes,
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Classes] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - Classes] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

func (h *Handler) vote(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Vote] Incoming request")

	classId := chi.URLParam(r, "id")

	p, err := h.pairingRepo.GetRandom(r.Context(), classId)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get random pairing. %v", err)
		// A nonexistent class (or one with no pairings) surfaces as a
		// not-found repo error -> 404, not a 500.
		render.Render(w, r, domain.ErrFromRepo(err))
		return
	}

	t1, err := h.torroRepo.Get(r.Context(), p.Torro1)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t2, err := h.torroRepo.Get(r.Context(), p.Torro2)
	if err != nil {
		logger.Error("[Handler - Vote] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t1.Pairing = p.Id
	t2.Pairing = p.Id

	// Look up the current user's voting streak for the small streak-indicator
	// pill (design prompt 13). This is purely decorative engagement UI, so an
	// anonymous session or a lookup failure just leaves it hidden rather than
	// failing the vote screen.
	var currentStreak int
	var streakAtRisk bool
	if userId := GetUserIDFromContext(r.Context()); userId != "" {
		if user, err := h.userRepo.Get(r.Context(), userId); err != nil {
			logger.Warn("[Handler - Vote] Couldn't get user for streak indicator. %v", err)
		} else {
			currentStreak = user.CurrentStreak
			today := time.Now().UTC().Format("2006-01-02")
			// LastVoteDate comes back from the DATE column via database/sql's
			// time.Time->string conversion, which yields an RFC3339 timestamp
			// (e.g. "2026-07-06T00:00:00Z"), not a bare "2006-01-02" date --
			// so compare by prefix rather than exact equality.
			streakAtRisk = currentStreak > 0 &&
				(user.LastVoteDate == nil || !strings.HasPrefix(*user.LastVoteDate, today))
		}
	}

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "vote.html", Content{
		Pairing:       p,
		Torrons:       []*domain.Torro{t1, t2},
		HX:            isHX(r),
		CurrentStreak: currentStreak,
		StreakAtRisk:  streakAtRisk,
	}); err != nil {
		logger.Error("[Handler - Vote] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - Vote] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

func (h *Handler) result(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Result] Incoming request")

	pairingId := chi.URLParam(r, "id")
	winnerId := r.URL.Query().Get("id")

	// isAdvent marks this submission as today's featured "advent daily duel"
	// vote (see the /advent page). It's still a normal vote through this same
	// codepath -- the flag only additionally records an AdventVotes row so
	// the user can't repeat today's featured duel, and changes what's
	// rendered afterwards (a "come back tomorrow" state instead of a fresh
	// random pairing).
	isAdvent := r.URL.Query().Get("advent") == "true"

	// Get user ID from context (set by UserMiddleware)
	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Warn("[Handler - Result] No user ID in context")
		// Continue without user tracking for backward compatibility
	}

	p, err := h.pairingRepo.Get(r.Context(), pairingId)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get pairing with ID %s. %v", pairingId, err)
		// A nonexistent/malformed pairing id surfaces as a not-found repo
		// error -> 404, not a 500.
		render.Render(w, r, domain.ErrFromRepo(err))
		return
	}

	// Validate that the winner ID matches one of the torros in the pairing
	if winnerId != p.Torro1 && winnerId != p.Torro2 {
		logger.Error("[Handler - Result] Invalid winner ID %s for pairing %s (expected %s or %s)",
			winnerId, pairingId, p.Torro1, p.Torro2)
		render.Render(w, r, domain.ErrBadRequest(
			fmt.Errorf("%s: Winner ID does not match pairing torros", domain.ValidationError)))
		return
	}

	// Look up the active campaign to tag this vote with, BEFORE opening the
	// transaction. GetActive uses the pooled *sql.DB (not tx); running it while
	// the transaction below holds its own pooled connection (and, now, FOR
	// UPDATE row locks) means each in-flight vote would hold one connection and
	// block waiting for a second — under concurrent votes that exhausts the
	// pool and deadlocks. This is a soft attach: voting isn't restricted to the
	// campaign window, so no active campaign just leaves CampaignId nil.
	var campaignIdPtr *string
	if campaign, err := h.campaignRepo.GetActive(r.Context()); err == nil {
		campaignIdPtr = &campaign.Id
	} else {
		logger.Debug("[Handler - Result] No active campaign to tag this vote with. %v", err)
	}

	// Start transaction to prevent race conditions in concurrent votes
	tx, err := h.db.Begin()
	if err != nil {
		logger.Error("[Handler - Result] Couldn't start transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	defer tx.Rollback() // Rollback if not committed

	// Read + lock both torró rows FOR UPDATE (GetTx takes a row lock) so
	// concurrent votes on the same torró serialize their read-modify-write of
	// "Rating" instead of clobbering each other (lost update). The two rows are
	// locked in a deterministic order (sorted by id) so two votes touching the
	// same pair can never deadlock by acquiring the locks in opposite orders.
	// t1/t2 keep their pairing meaning (t1 == p.Torro1) regardless of lock order.
	var t1, t2 *domain.Torro
	if p.Torro1 <= p.Torro2 {
		if t1, err = h.torroRepo.GetTx(tx, r.Context(), p.Torro1); err == nil {
			t2, err = h.torroRepo.GetTx(tx, r.Context(), p.Torro2)
		}
	} else {
		if t2, err = h.torroRepo.GetTx(tx, r.Context(), p.Torro2); err == nil {
			t1, err = h.torroRepo.GetTx(tx, r.Context(), p.Torro1)
		}
	}
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	// Calculate new ELO ratings based on match result
	new1, new2 := UpdateRatings(t1.Rating, t2.Rating, winnerId == t1.Id, K)

	// Prepare user ID pointer for result record (nullable)
	var userIdPtr *string
	if userId != "" {
		userIdPtr = &userId
	}

	// Create result record within transaction
	_, err = h.resultRepo.CreateTx(tx, r.Context(), &domain.Result{
		Pairing:    pairingId,
		Rat1Bef:    t1.Rating,
		Rat2Bef:    t2.Rating,
		Winner:     winnerId,
		Rat1Aft:    new1,
		Rat2Aft:    new2,
		UserId:     userIdPtr, // Track which user cast this vote
		CampaignId: campaignIdPtr,
	})
	if err != nil {
		logger.Error("[Handler - Result] Couldn't create result. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	// Update user vote count if user tracking is enabled
	if userId != "" && p.Class != "" {
		if err := h.userRepo.IncrementVoteCountTx(tx, r.Context(), userId, p.Class); err != nil {
			logger.Error("[Handler - Result] Couldn't increment user vote count. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}

		// Update the user's voting streak (any vote in any class counts).
		// This is a general engagement metric, independent of per-class counts.
		if err := h.userRepo.UpdateStreakTx(tx, r.Context(), userId); err != nil {
			logger.Error("[Handler - Result] Couldn't update user streak. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}

		// Calculate and update personalized ELO ratings for this user
		// Get or create user ELO snapshots for both torrons
		userElo1, err := h.userEloRepo.GetOrCreateTx(tx, r.Context(), userId, t1.Id)
		if err != nil {
			logger.Error("[Handler - Result] Couldn't get user ELO for torron 1. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}

		userElo2, err := h.userEloRepo.GetOrCreateTx(tx, r.Context(), userId, t2.Id)
		if err != nil {
			logger.Error("[Handler - Result] Couldn't get user ELO for torron 2. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}

		// Calculate new personalized ELO ratings (same formula as global)
		userNew1, userNew2 := UpdateRatings(userElo1.Rating, userElo2.Rating, winnerId == t1.Id, K)

		// Update user ELO snapshots
		userElo1.Rating = userNew1
		userElo1.VoteCount++
		if _, err := h.userEloRepo.UpdateTx(tx, r.Context(), userElo1); err != nil {
			logger.Error("[Handler - Result] Couldn't update user ELO for torron 1. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}

		userElo2.Rating = userNew2
		userElo2.VoteCount++
		if _, err := h.userEloRepo.UpdateTx(tx, r.Context(), userElo2); err != nil {
			logger.Error("[Handler - Result] Couldn't update user ELO for torron 2. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}
	}

	// Update global ratings within transaction
	_, err = h.torroRepo.UpdateTx(tx, r.Context(), t1.Id, new1)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't update rating: %s. %v", t1.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	_, err = h.torroRepo.UpdateTx(tx, r.Context(), t2.Id, new2)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't update rating: %s. %v", t2.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	// If this is today's featured advent duel, record it within the same
	// transaction so the whole vote (ELO update included) rolls back if the
	// user has somehow already completed today's duel (unique constraint on
	// UserId+VoteDate).
	if isAdvent {
		if userId == "" {
			logger.Error("[Handler - Result] Advent vote submitted without a user ID")
			render.Render(w, r, domain.ErrBadRequest(
				fmt.Errorf("%s: A user is required to record an advent vote", domain.ValidationError)))
			return
		}

		voteDate := time.Now().UTC().Format("2006-01-02")
		if _, err := h.adventVoteRepo.CreateTx(tx, r.Context(), &domain.AdventVote{
			UserId:    userId,
			VoteDate:  voteDate,
			PairingId: pairingId,
		}); err != nil {
			logger.Error("[Handler - Result] Couldn't record advent vote. %v", err)
			// A repeat of today's featured duel hits the UNIQUE(UserId,VoteDate)
			// constraint -> duplicate-key -> 409 Conflict, not a 500.
			render.Render(w, r, domain.ErrFromRepo(err))
			return
		}
	}

	// Commit transaction (makes all changes visible atomically)
	if err := tx.Commit(); err != nil {
		logger.Error("[Handler - Result] Couldn't commit transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	// Advent duels are once-per-day: instead of serving a fresh random
	// pairing like the normal voting flow, show the "come back tomorrow"
	// state.
	if isAdvent {
		buf := h.bpool.Get()
		defer h.bpool.Put(buf)

		if err := h.template.ExecuteTemplate(buf, "advent.html", AdventContent{
			HX:             isHX(r),
			CampaignActive: true,
			AlreadyVoted:   true,
		}); err != nil {
			logger.Error("[Handler - Result] Couldn't execute advent template. %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		buf.WriteTo(w)
		return
	}

	newP, err := h.pairingRepo.GetRandom(r.Context(), p.Class)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get random pairing. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt1, err := h.torroRepo.Get(r.Context(), newP.Torro1)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt2, err := h.torroRepo.Get(r.Context(), newP.Torro2)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	newt1.Pairing = newP.Id
	newt2.Pairing = newP.Id

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "pairing.html", Content{
		Pairing: newP,
		Torrons: []*domain.Torro{newt1, newt2},
		HX:      isHX(r),
	}); err != nil {
		logger.Error("[Handler - Result] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - Result] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

func isHX(r *http.Request) bool {
	if r.Header.Get("HX-Request") == "true" {
		return true
	}
	return false
}
