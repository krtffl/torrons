package http

import (
	"database/sql"
	"fmt"
	"html/template"
	"net/http"

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
}

type Handler struct {
	db           *sql.DB
	template     *template.Template
	bpool        *bpool.BufferPool
	pairingRepo  domain.PairingRepo
	torroRepo    domain.TorroRepo
	classRepo    domain.ClassRepo
	resultRepo   domain.ResultRepo
	userRepo     domain.UserRepo
	userEloRepo  domain.UserEloSnapshotRepo
	campaignRepo domain.CampaignRepo
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
) *Handler {
	tmpls, err := template.New("").ParseFS(torrons.Public, "public/templates/*.html")
	if err != nil {
		logger.Fatal("[Handler] - Failed to parse templates. %v", err)
	}

	return &Handler{
		db:           db,
		template:     tmpls,
		bpool:        bpool,
		pairingRepo:  pairingRep,
		torroRepo:    torroRepo,
		classRepo:    classRepo,
		resultRepo:   resultRepo,
		userRepo:     userRepo,
		userEloRepo:  userEloRepo,
		campaignRepo: campaignRepo,
	}
}

func (h *Handler) index(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Index] Incoming request")

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "index.html", Content{}); err != nil {
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
		render.Render(w, r, domain.ErrInternal(err))
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

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "vote.html", Content{
		Pairing: p,
		Torrons: []*domain.Torro{t1, t2},
		HX:      isHX(r),
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

	// Get user ID from context (set by UserMiddleware)
	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Warn("[Handler - Result] No user ID in context")
		// Continue without user tracking for backward compatibility
	}

	p, err := h.pairingRepo.Get(r.Context(), pairingId)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get pairing with ID %s. %v", pairingId, err)
		render.Render(w, r, domain.ErrInternal(err))
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

	// Start transaction to prevent race conditions in concurrent votes
	tx, err := h.db.Begin()
	if err != nil {
		logger.Error("[Handler - Result] Couldn't start transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	defer tx.Rollback() // Rollback if not committed

	// Get both torros within transaction (ensures consistent read)
	t1, err := h.torroRepo.GetTx(tx, r.Context(), p.Torro1)
	if err != nil {
		logger.Error("[Handler - Result] Couldn't get torro. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	t2, err := h.torroRepo.GetTx(tx, r.Context(), p.Torro2)
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
		Pairing: pairingId,
		Rat1Bef: t1.Rating,
		Rat2Bef: t2.Rating,
		Winner:  winnerId,
		Rat1Aft: new1,
		Rat2Aft: new2,
		UserId:  userIdPtr, // Track which user cast this vote
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

	// Commit transaction (makes all changes visible atomically)
	if err := tx.Commit(); err != nil {
		logger.Error("[Handler - Result] Couldn't commit transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
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
