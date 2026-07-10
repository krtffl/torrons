package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/krtffl/torro/internal/logger"
	"github.com/krtffl/torro/internal/sharecard"
)

// RevealContent holds data for the /reveal page template. It mirrors
// sharecard.RevealData field-for-field (plus HX) on purpose: both the
// page and the PNG card are built from the exact same revealCardData call
// below, so they can never show different numbers - same rationale as
// WrappedContent/wrappedCardData.
type RevealContent struct {
	HX bool

	HasEnoughVotes bool
	VotesRemaining int
	TotalVotes     int

	HasClearFavorite bool
	Percentile       int

	PersonaBadge   string
	PersonaTagline string

	TopTorroName      string
	TopTorroTag       string
	TopTorroVotesCast int
	TopTorroWins      int
	TopTorroLosses    int
}

// personaCopyByClass is the approved "torró personality" copy table (design
// prompt 14, signed off verbatim), keyed by the voting ARENA the user
// voted in the most (Pairings.Class - see domain.PersonaStats.TopClassId's
// doc comment for why it's the arena and not Torrons.Class). taglineSuffix
// completes "Ets del X% que <suffix>", matching this variant's own
// personaTagline prop in docs/design-deliverables/Torrorendum Story
// Card.dc.html.
var personaCopyByClass = map[string]struct {
	badge         string
	taglineSuffix string
}{
	"3": {"ELS GOLAFRES DE XOCOLATA", "no pot resistir-se a un torró de xocolata."},
	"1": {"ELS TRADICIONALISTES", "no necessita res més que ametlla, mel i tradició."},
	"2": {"ELS EXPLORADORS", "sempre vol tastar el sabor que ningú ha provat encara."},
	"4": {"ELS ATREVITS", "prefereix els sabors més atrevits, signats per un xef."},
	"5": {"ELS ÀRBITRES", "no es conforma amb una categoria: vol saber qui guanya de veritat."},
}

const (
	// equilibratsBadge/equilibratsTagline are the approved copy for the
	// tie/no-clear-favorite persona. Unlike the five classed personas
	// above, it never completes an "Ets del X% que..." clause (there is no
	// percentile on an exact tie - see domain.PersonaStats.HasClearFavorite),
	// so it stands as its own sentence. The approved tagline text itself
	// is reproduced with no wording changed - only capitalized, since it's
	// sentence-initial here instead of a "que ..." continuation.
	equilibratsBadge   = "ELS EQUILIBRATS"
	equilibratsTagline = "Reparteix els vots per igual entre totes les categories."
)

// arenaTagForClass returns the same per-arena monospace tag copy already
// shown next to each category name in public/templates/classes.html (see
// its "class" template block's arena-tag span) - kept as a single shared
// Go implementation, rather than re-typed differently here and (per a
// parallel branch) on the personal stats page, so the two pages can never
// drift out of sync on this copy.
func arenaTagForClass(classId string) string {
	switch classId {
	case "1":
		return "L'ORIGINAL"
	case "2":
		return "EN TENDÈNCIA"
	case "3":
		return "PER ALS GOLAFRES"
	case "4":
		return "EDICIÓ LIMITADA"
	case "5":
		return "EL REPTE DEFINITIU"
	default:
		return ""
	}
}

// reveal handles GET /reveal: the "torró personality reveal" page (design
// prompt 14). Gated behind the same minimum-vote threshold as Wrapped/the
// Global leaderboard; below it, the page renders an honest "not unlocked
// yet" state instead of fabricating a persona from too little data.
func (h *Handler) reveal(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - Reveal] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - Reveal] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	data, err := h.revealCardData(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - Reveal] Couldn't fetch reveal data. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	content := revealContentFromCardData(data)
	content.HX = isHX(r)

	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "reveal.html", content); err != nil {
		logger.Error("[Handler - Reveal] Couldn't execute template. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	buf.WriteTo(w)
}

// revealCard renders and serves the personal reveal 1080x1920 PNG "story"
// card, using the exact same data as the reveal page above.
func (h *Handler) revealCard(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - RevealCard] Incoming request")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - RevealCard] No user ID in context")
		http.Error(w, "User not found", http.StatusUnauthorized)
		return
	}

	data, err := h.revealCardData(r.Context(), userId)
	if err != nil {
		logger.Error("[Handler - RevealCard] Couldn't fetch reveal data. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Cap concurrent renders (each allocates a large RGBA and pegs a core):
	// shed load with 503 rather than piling up when the cap is saturated.
	if !sharecard.TryAcquireRenderSlot(r.Context()) {
		logger.Warn("[Handler - RevealCard] Render slots saturated; shedding request.")
		w.Header().Set("Retry-After", "1")
		http.Error(w, "Service Unavailable", http.StatusServiceUnavailable)
		return
	}
	defer sharecard.ReleaseRenderSlot()

	png, err := sharecard.RenderReveal(data)
	if err != nil {
		logger.Error("[Handler - RevealCard] Couldn't render card. %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Personalized per-user content: never cache/share across users, same
	// precedent as shareCard/wrappedCard.
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "private, no-store")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(png); err != nil {
		logger.Error("[Handler - RevealCard] Couldn't write response. %v", err)
	}
}

// revealCardData assembles the single canonical sharecard.RevealData value
// for one user, shared by both reveal (HTML) and revealCard (PNG) so their
// data-fetching logic never has to be duplicated or drift apart.
func (h *Handler) revealCardData(ctx context.Context, userId string) (sharecard.RevealData, error) {
	user, err := h.userRepo.Get(ctx, userId)
	if err != nil {
		return sharecard.RevealData{}, err
	}

	minVotes := getMinVotesForClass(embedDefaultClassId) // "5" - Global
	if user.VoteCount < minVotes {
		return sharecard.RevealData{
			HasEnoughVotes: false,
			VotesRemaining: minVotes - user.VoteCount,
			TotalVotes:     user.VoteCount,
		}, nil
	}

	stats, err := h.personaRepo.Stats(ctx, userId, minVotes)
	if err != nil {
		return sharecard.RevealData{}, err
	}

	data := sharecard.RevealData{
		HasEnoughVotes:    true,
		TotalVotes:        stats.TotalVotes,
		HasClearFavorite:  stats.HasClearFavorite,
		TopTorroName:      stats.TopTorroName,
		TopTorroVotesCast: stats.TopTorroVotesCast,
		TopTorroWins:      stats.TopTorroWins,
		TopTorroLosses:    stats.TopTorroLosses,
	}

	if stats.HasClearFavorite {
		copy := personaCopyByClass[stats.TopClassId]
		data.Percentile = stats.Percentile
		data.PersonaBadge = copy.badge
		data.PersonaTagline = fmt.Sprintf("que %s", copy.taglineSuffix)

		if stats.TopClassName != "" {
			data.TopTorroTag = fmt.Sprintf("%s · %s", strings.ToUpper(stats.TopClassName), arenaTagForClass(stats.TopClassId))
		}
	} else {
		data.PersonaBadge = equilibratsBadge
		data.PersonaTagline = equilibratsTagline

		// On a tie there's no arena persona to tag the torró card with, so
		// fall back to the resolved torró's own real Torrons.Class (unlike
		// the clear-favorite branch above, which always tags with the
		// ARENA's own class - see arenaTagForClass's doc comment and the
		// "GLOBAL · EL REPTE DEFINITIU" example in design prompt 14, where
		// the arena and the torró's real class can differ).
		if stats.TopTorroId != "" {
			if torro, err := h.torroRepo.Get(ctx, stats.TopTorroId); err != nil {
				logger.Warn("[Handler - Reveal] Couldn't resolve top torró's own class for tag copy. %v", err)
			} else if className := h.classNameById(ctx, torro.Class); className != "" {
				data.TopTorroTag = fmt.Sprintf("%s · %s", strings.ToUpper(className), arenaTagForClass(torro.Class))
			}
		}
	}

	return data, nil
}

// classNameById looks up a single class's Name. domain.ClassRepo only
// exposes List (the catalog is a fixed 5 rows), so listing and filtering
// here is simpler than adding a new repo method just for this one lookup.
func (h *Handler) classNameById(ctx context.Context, classId string) string {
	classes, err := h.classRepo.List(ctx)
	if err != nil {
		logger.Warn("[Handler - Reveal] Couldn't list classes for name lookup. %v", err)
		return ""
	}
	for _, c := range classes {
		if c.Id == classId {
			return c.Name
		}
	}
	return ""
}

// revealContentFromCardData copies sharecard.RevealData's fields into the
// template-facing RevealContent shape. Kept as a pure field copy (no extra
// derivation) so reveal.html and the PNG card are guaranteed to show the
// exact same numbers and copy.
func revealContentFromCardData(data sharecard.RevealData) RevealContent {
	return RevealContent{
		HasEnoughVotes:    data.HasEnoughVotes,
		VotesRemaining:    data.VotesRemaining,
		TotalVotes:        data.TotalVotes,
		HasClearFavorite:  data.HasClearFavorite,
		Percentile:        data.Percentile,
		PersonaBadge:      data.PersonaBadge,
		PersonaTagline:    data.PersonaTagline,
		TopTorroName:      data.TopTorroName,
		TopTorroTag:       data.TopTorroTag,
		TopTorroVotesCast: data.TopTorroVotesCast,
		TopTorroWins:      data.TopTorroWins,
		TopTorroLosses:    data.TopTorroLosses,
	}
}
