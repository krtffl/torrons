package http

import (
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"math/big"
	"math/bits"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"

	"github.com/krtffl/torro/internal/domain"
	"github.com/krtffl/torro/internal/logger"
)

// Phase 2 - The knockout.
//
// This file is the HTTP surface for the single-elimination bracket that
// sits on top of Phase 1's open season. It is a deliberately different
// mechanic (see internal/domain/bracket.go): one vote per user per match,
// a round is resolved by tally rather than an ELO nudge, and none of this
// ever touches Torro.Rating. Do not reuse this file's helpers from the
// Phase 1 pairing/vote/result code, and vice versa.

// BracketTorroView is a torró as displayed within a bracket match or as
// bracket champion.
type BracketTorroView struct {
	Id    string
	Name  string
	Image string
	Seed  int
}

// BracketMatchView is a single match ready for template rendering.
// Torro1Won / Torro2Won are plain bools (rather than leaving the template
// to compare IDs against a possibly-nil Winner pointer) so bracket.html
// doesn't need any nil-safety logic to render a winner badge.
type BracketMatchView struct {
	Id        string
	Round     int
	Slot      int
	Status    string
	IsBye     bool
	Decided   bool
	Torro1    BracketTorroView
	Torro2    *BracketTorroView // nil for a bye
	Torro1Won bool
	Torro2Won bool
}

// BracketRoundView groups matches by round for the overview page.
type BracketRoundView struct {
	Round     int
	IsFinal   bool
	IsCurrent bool
	Matches   []BracketMatchView
}

// BracketOverviewContent holds data for the bracket overview page.
type BracketOverviewContent struct {
	HX            bool
	ClassId       string
	ClassName     string
	BracketExists bool
	Bracket       *domain.Bracket
	Rounds        []BracketRoundView
	Champion      *BracketTorroView
	TotalRounds   int
}

// BracketVoteContent holds data for the bracket voting card.
type BracketVoteContent struct {
	HX            bool
	ClassId       string
	BracketExists bool
	Completed     bool
	ChampionName  string
	NoMoreMatches bool
	Match         *BracketMatchView
}

// bracketOverview handles GET /bracket/{classId}: current round, every
// match with its status, past-round results and the champion once decided.
func (h *Handler) bracketOverview(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - BracketOverview] Incoming request")

	classId := chi.URLParam(r, "classId")
	ctx := r.Context()

	className := classId
	if classes, err := h.classRepo.List(ctx); err == nil {
		for _, c := range classes {
			if c.Id == classId {
				className = c.Name
				break
			}
		}
	}

	content := BracketOverviewContent{
		HX:        isHX(r),
		ClassId:   classId,
		ClassName: className,
	}

	bracket, err := h.bracketRepo.GetLatestByClass(ctx, classId)
	if err != nil {
		// No bracket for this class yet - not an error, just an empty state.
		h.renderBracketOverview(w, r, content)
		return
	}
	content.BracketExists = true
	content.Bracket = bracket
	content.TotalRounds = bits.Len(uint(bracket.Size)) - 1

	entries, err := h.bracketRepo.ListEntries(ctx, bracket.Id)
	if err != nil {
		logger.Error("[Handler - BracketOverview] Couldn't list entries for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	seedByTorro := seedMap(entries)

	matches, err := h.bracketRepo.ListMatches(ctx, bracket.Id)
	if err != nil {
		logger.Error("[Handler - BracketOverview] Couldn't list matches for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	getTorro := h.torroFetcher(ctx)

	roundsByNumber := make(map[int][]BracketMatchView)
	maxRound := 0
	for _, m := range matches {
		view, err := buildMatchView(m, seedByTorro, getTorro)
		if err != nil {
			logger.Error("[Handler - BracketOverview] Couldn't build match view for %s. %v", m.Id, err)
			render.Render(w, r, domain.ErrInternal(err))
			return
		}
		roundsByNumber[m.Round] = append(roundsByNumber[m.Round], view)
		if m.Round > maxRound {
			maxRound = m.Round
		}
	}

	for round := 1; round <= maxRound; round++ {
		content.Rounds = append(content.Rounds, BracketRoundView{
			Round:     round,
			IsFinal:   round == content.TotalRounds,
			IsCurrent: round == bracket.CurrentRound && bracket.Status == domain.BracketStatusInProgress,
			Matches:   roundsByNumber[round],
		})
	}

	if bracket.ChampionId != nil {
		if champ, err := getTorro(*bracket.ChampionId); err == nil {
			content.Champion = &BracketTorroView{
				Id:    champ.Id,
				Name:  champ.Name,
				Image: champ.Image,
				Seed:  seedByTorro[champ.Id],
			}
		}
	}

	h.renderBracketOverview(w, r, content)
}

func (h *Handler) renderBracketOverview(w http.ResponseWriter, r *http.Request, content BracketOverviewContent) {
	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "bracket.html", content); err != nil {
		logger.Error("[Handler - BracketOverview] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - BracketOverview] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

// bracketVote handles GET /bracket/{classId}/vote: serve one of the
// current round's still-open matches, picked at random among the ones the
// viewer hasn't voted on yet, mirroring PairingRepo.GetRandom's approach
// to spreading votes across concurrent matches.
func (h *Handler) bracketVote(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - BracketVote] Incoming request")

	classId := chi.URLParam(r, "classId")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Error("[Handler - BracketVote] No user ID in context")
		render.Render(w, r, domain.ErrInternal(fmt.Errorf("%s: missing user context", domain.ValidationError)))
		return
	}

	h.serveBracketVoteCard(w, r, classId, userId)
}

// serveBracketVoteCard renders the current voting state for a class: a
// random still-open match, or a message if none are left for this viewer,
// or the champion if the bracket already concluded.
func (h *Handler) serveBracketVoteCard(w http.ResponseWriter, r *http.Request, classId string, userId string) {
	ctx := r.Context()
	content := BracketVoteContent{HX: isHX(r), ClassId: classId}

	bracket, err := h.bracketRepo.GetLatestByClass(ctx, classId)
	if err != nil {
		h.renderBracketVotePage(w, r, content)
		return
	}
	content.BracketExists = true

	if bracket.Status == domain.BracketStatusCompleted {
		content.Completed = true
		if bracket.ChampionId != nil {
			if champ, err := h.torroRepo.Get(ctx, *bracket.ChampionId); err == nil {
				content.ChampionName = champ.Name
			}
		}
		h.renderBracketVotePage(w, r, content)
		return
	}

	openMatches, err := h.bracketRepo.ListOpenMatchesForUser(ctx, bracket.Id, bracket.CurrentRound, userId)
	if err != nil {
		logger.Error("[Handler - BracketVote] Couldn't list open matches for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	match, err := pickRandomMatch(openMatches)
	if err != nil {
		logger.Error("[Handler - BracketVote] Couldn't pick random match. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	if match == nil {
		content.NoMoreMatches = true
		h.renderBracketVotePage(w, r, content)
		return
	}

	entries, err := h.bracketRepo.ListEntries(ctx, bracket.Id)
	if err != nil {
		logger.Error("[Handler - BracketVote] Couldn't list entries for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	view, err := buildMatchView(match, seedMap(entries), h.torroFetcher(ctx))
	if err != nil {
		logger.Error("[Handler - BracketVote] Couldn't build match view. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	content.Match = &view
	h.renderBracketVotePage(w, r, content)
}

func (h *Handler) renderBracketVotePage(w http.ResponseWriter, r *http.Request, content BracketVoteContent) {
	buf := h.bpool.Get()
	defer h.bpool.Put(buf)

	if err := h.template.ExecuteTemplate(buf, "bracket-vote-page", content); err != nil {
		logger.Error("[Handler - BracketVote] Couldn't execute template. %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		if execErr := h.template.ExecuteTemplate(w, "error.html", Content{}); execErr != nil {
			logger.Error("[Handler - BracketVote] Failed to render error page. %v", execErr)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
		return
	}

	buf.WriteTo(w)
}

// bracketMatchVote handles POST /bracket/match/{matchId}/vote?winner={id}.
// Casts a user's vote for a match (at most one vote per user per match,
// enforced by a DB unique constraint) and, in the same transaction, checks
// whether that vote closed out the round - if so it tallies every match in
// the round and advances the bracket. This mirrors handler.go's `result`
// function: a read-then-write that needs transactional consistency.
func (h *Handler) bracketMatchVote(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - BracketMatchVote] Incoming request")

	matchId := chi.URLParam(r, "matchId")
	winnerId := r.URL.Query().Get("winner")

	userId := GetUserIDFromContext(r.Context())
	if userId == "" {
		logger.Warn("[Handler - BracketMatchVote] No user ID in context")
		render.Render(w, r, domain.ErrBadRequest(fmt.Errorf("%s: missing user context", domain.ValidationError)))
		return
	}

	if winnerId == "" {
		render.Render(w, r, domain.ErrBadRequest(fmt.Errorf("%s: missing winner query parameter", domain.ValidationError)))
		return
	}

	ctx := r.Context()

	tx, err := h.db.Begin()
	if err != nil {
		logger.Error("[Handler - BracketMatchVote] Couldn't start transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	defer tx.Rollback() // Rollback if not committed

	match, err := h.bracketRepo.GetMatchTx(tx, ctx, matchId)
	if err != nil {
		logger.Error("[Handler - BracketMatchVote] Couldn't get match %s. %v", matchId, err)
		renderBracketError(w, r, err)
		return
	}

	if match.Status != domain.BracketMatchStatusPending {
		render.Render(w, r, domain.ErrBadRequest(
			fmt.Errorf("%s: this match is no longer open for voting", domain.ValidationError)))
		return
	}

	if winnerId != match.Torro1Id && (match.Torro2Id == nil || winnerId != *match.Torro2Id) {
		render.Render(w, r, domain.ErrBadRequest(
			fmt.Errorf("%s: winner does not match either competitor in this match", domain.ValidationError)))
		return
	}

	bracket, err := h.bracketRepo.GetTx(tx, ctx, match.BracketId)
	if err != nil {
		logger.Error("[Handler - BracketMatchVote] Couldn't get bracket %s. %v", match.BracketId, err)
		renderBracketError(w, r, err)
		return
	}

	_, err = h.bracketRepo.CreateVoteTx(tx, ctx, &domain.BracketMatchVote{
		MatchId:  matchId,
		UserId:   userId,
		TorronId: winnerId,
	})
	if err != nil {
		if strings.Contains(err.Error(), string(domain.DuplicateKeyError)) {
			logger.Warn("[Handler - BracketMatchVote] User %s already voted on match %s", userId, matchId)
			render.Render(w, r, domain.ErrBadRequest(
				fmt.Errorf("%s: you have already voted on this match", domain.ValidationError)))
		} else {
			logger.Error("[Handler - BracketMatchVote] Couldn't record vote. %v", err)
			render.Render(w, r, domain.ErrInternal(err))
		}
		return
	}

	// Lazy check-and-advance: judgment call worth double-checking.
	//
	// There is no fixed voter roster for a bracket match (anyone can vote
	// once, but nobody is required to), so "the round is fully voted"
	// can't mean 100% turnout the way Phase 1's min-vote thresholds do.
	// We treat a round as fully voted once every match in it has received
	// at least one vote (byes don't count, they're already decided). This
	// keeps the knockout moving without a cron job, at the cost of a
	// single slow/ignored match blocking the whole round - that's exactly
	// what the explicit force-advance endpoint is for.
	if err := h.checkAndAdvanceIfRoundFullyVoted(tx, ctx, bracket); err != nil {
		logger.Error("[Handler - BracketMatchVote] Couldn't check/advance round for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("[Handler - BracketMatchVote] Couldn't commit transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	h.serveBracketVoteCard(w, r, bracket.ClassId, userId)
}

// checkAndAdvanceIfRoundFullyVoted resolves and cascades the bracket's
// current round if every pending match in it has at least one vote. It is
// a no-op otherwise.
func (h *Handler) checkAndAdvanceIfRoundFullyVoted(tx *sql.Tx, ctx context.Context, bracket *domain.Bracket) error {
	roundMatches, err := h.bracketRepo.ListMatchesByRoundTx(tx, ctx, bracket.Id, bracket.CurrentRound)
	if err != nil {
		return err
	}

	for _, m := range roundMatches {
		if m.Status != domain.BracketMatchStatusPending {
			continue
		}

		votes, err := h.bracketRepo.CountVotesByTorronTx(tx, ctx, m.Id)
		if err != nil {
			return err
		}

		if totalVotes(votes) == 0 {
			// At least one match hasn't been touched yet - round isn't
			// fully voted, nothing to advance.
			return nil
		}
	}

	entries, err := h.bracketRepo.ListEntriesTx(tx, ctx, bracket.Id)
	if err != nil {
		return err
	}
	seedByTorro := seedMap(entries)

	if err := h.resolvePendingMatchesInRound(tx, ctx, bracket, seedByTorro); err != nil {
		return err
	}

	previousRound := bracket.CurrentRound
	if _, err := h.cascadeAdvance(tx, ctx, bracket); err != nil {
		return err
	}

	logger.Info("[Handler - BracketMatchVote] Round %d of bracket %s fully voted, now at round %d (status=%s)",
		previousRound, bracket.Id, bracket.CurrentRound, bracket.Status)

	return nil
}

// resolvePendingMatchesInRound tallies votes for every still-pending match
// in bracket.CurrentRound and sets its winner: whichever torró has more
// votes advances; a tie (including 0-0, which happens on a forced advance
// of an untouched match) is broken by the lower original seed number, i.e.
// the torró Phase 1 rated higher advances. This tie-break is a judgment
// call worth double-checking against the intended product behaviour.
func (h *Handler) resolvePendingMatchesInRound(tx *sql.Tx, ctx context.Context, bracket *domain.Bracket, seedByTorro map[string]int) error {
	matches, err := h.bracketRepo.ListMatchesByRoundTx(tx, ctx, bracket.Id, bracket.CurrentRound)
	if err != nil {
		return err
	}

	for _, m := range matches {
		if m.Status != domain.BracketMatchStatusPending {
			continue
		}

		votes, err := h.bracketRepo.CountVotesByTorronTx(tx, ctx, m.Id)
		if err != nil {
			return err
		}

		winnerId := decideMatchWinner(m, votes, seedByTorro)
		if err := h.bracketRepo.SetMatchWinnerTx(tx, ctx, m.Id, winnerId); err != nil {
			return err
		}
	}

	return nil
}

// cascadeAdvance assumes every match that exists in bracket.CurrentRound
// has already been decided (Status == completed). It builds the next
// round from the current round's winners and keeps cascading through any
// round that turns out to be pure byes (which can happen when a class has
// far fewer torrons than the requested bracket size), stopping either when
// a round has at least one match that genuinely needs votes, or when the
// bracket is completed. Returns whether the bracket was completed.
func (h *Handler) cascadeAdvance(tx *sql.Tx, ctx context.Context, bracket *domain.Bracket) (bool, error) {
	for {
		roundMatches, err := h.bracketRepo.ListMatchesByRoundTx(tx, ctx, bracket.Id, bracket.CurrentRound)
		if err != nil {
			return false, err
		}

		for _, m := range roundMatches {
			if m.Status == domain.BracketMatchStatusPending {
				// Round genuinely isn't decided yet; nothing to advance.
				return false, nil
			}
		}

		// Number of matches this round would have if every seed showed up.
		slotCount := bracket.Size >> uint(bracket.CurrentRound)
		if slotCount <= 1 {
			// This round IS the Gran Final.
			if len(roundMatches) != 1 || roundMatches[0].WinnerId == nil {
				return false, fmt.Errorf(
					"%s: bracket %s has no resolvable final match in round %d",
					domain.ValidationError, bracket.Id, bracket.CurrentRound)
			}

			championId := *roundMatches[0].WinnerId
			if err := h.bracketRepo.CompleteTx(tx, ctx, bracket.Id, championId); err != nil {
				return false, err
			}
			bracket.Status = domain.BracketStatusCompleted
			bracket.ChampionId = &championId
			return true, nil
		}

		winners := make([]*string, slotCount)
		for _, m := range roundMatches {
			if m.Slot >= 0 && m.Slot < slotCount && m.WinnerId != nil {
				winnerId := *m.WinnerId
				winners[m.Slot] = &winnerId
			}
		}

		nextRound := bracket.CurrentRound + 1
		nextSlotCount := slotCount / 2
		createdPending := false

		for slot := 0; slot < nextSlotCount; slot++ {
			a := winners[2*slot]
			b := winners[2*slot+1]

			if a == nil && b == nil {
				// Both sub-brackets were empty (only possible when a class
				// has far fewer torrons than the requested bracket size) -
				// propagate the gap, no match to create for this slot.
				continue
			}

			match := &domain.BracketMatch{
				BracketId: bracket.Id,
				Round:     nextRound,
				Slot:      slot,
			}

			if a != nil && b != nil {
				match.Torro1Id = *a
				torro2Id := *b
				match.Torro2Id = &torro2Id
				match.Status = domain.BracketMatchStatusPending
				createdPending = true
			} else {
				winnerId := a
				if winnerId == nil {
					winnerId = b
				}
				match.Torro1Id = *winnerId
				match.Status = domain.BracketMatchStatusCompleted
				w := *winnerId
				match.WinnerId = &w
			}

			if _, err := h.bracketRepo.CreateMatchTx(tx, ctx, match); err != nil {
				return false, err
			}
		}

		if err := h.bracketRepo.UpdateRoundTx(tx, ctx, bracket.Id, nextRound); err != nil {
			return false, err
		}
		bracket.CurrentRound = nextRound

		if createdPending {
			return false, nil
		}
		// Otherwise the new round is entirely byes (or gaps): loop again
		// to cascade straight through it.
	}
}

// bracketCreate handles POST /bracket/{classId}/create?size={n}.
// Gated by Handler.RequireAdminToken - see its route registration in server.go.
func (h *Handler) bracketCreate(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - BracketCreate] Incoming request")

	classId := chi.URLParam(r, "classId")

	size := domain.DefaultBracketSize
	if sizeParam := r.URL.Query().Get("size"); sizeParam != "" {
		parsed, err := strconv.Atoi(sizeParam)
		if err != nil {
			render.Render(w, r, domain.ErrBadRequest(
				fmt.Errorf("%s: size must be an integer", domain.ValidationError)))
			return
		}
		size = parsed
	}

	bracket, err := h.seedAndCreateBracket(r.Context(), classId, size)
	if err != nil {
		logger.Error("[Handler - BracketCreate] Couldn't create bracket for class %s. %v", classId, err)
		renderBracketError(w, r, err)
		return
	}

	render.Status(r, http.StatusCreated)
	render.JSON(w, r, bracket)
}

// seedAndCreateBracket seeds the top-N torrons of a class by Phase 1
// rating (N = min(size, active torrons in class)) and generates round-1
// matches using standard single-elimination seeding (1v8, 4v5, 2v7, 3v6
// for a field of 8, generalized to any power-of-two size). If N isn't a
// power of two, the missing top seeds are byes that auto-advance.
func (h *Handler) seedAndCreateBracket(ctx context.Context, classId string, size int) (*domain.Bracket, error) {
	if !isPowerOfTwo(size) {
		return nil, fmt.Errorf("%s: bracket size must be a power of two (got %d)", domain.ValidationError, size)
	}

	campaign, err := h.campaignRepo.GetActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: no active campaign to attach a bracket to", domain.ValidationError)
	}

	if existing, err := h.bracketRepo.GetByCampaignAndClass(ctx, campaign.Id, classId); err == nil && existing != nil {
		return nil, fmt.Errorf(
			"%s: a bracket already exists for class %s in the active campaign", domain.ValidationError, classId)
	}

	topTorrons, err := h.torroRepo.TopNByClass(ctx, classId, size)
	if err != nil {
		return nil, err
	}
	if len(topTorrons) < 2 {
		return nil, fmt.Errorf(
			"%s: class %s needs at least 2 active torrons to start a bracket", domain.ValidationError, classId)
	}

	tx, err := h.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	bracket := &domain.Bracket{
		CampaignId:   campaign.Id,
		ClassId:      classId,
		Size:         size,
		CurrentRound: 1,
	}
	bracket, err = h.bracketRepo.CreateTx(tx, ctx, bracket)
	if err != nil {
		return nil, err
	}

	for i, t := range topTorrons {
		_, err := h.bracketRepo.CreateEntryTx(tx, ctx, &domain.BracketEntry{
			BracketId:  bracket.Id,
			TorronId:   t.Id,
			Seed:       i + 1,
			SeedRating: t.Rating,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := h.createRoundOneMatches(tx, ctx, bracket, topTorrons); err != nil {
		return nil, err
	}

	// A tiny/sparse field can resolve entirely via byes before a single
	// vote is cast; cascade straight through that case too.
	if _, err := h.cascadeAdvance(tx, ctx, bracket); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return bracket, nil
}

// createRoundOneMatches lays out round 1 using the standard bracket seed
// order. topTorrons must already be sorted by seed (1..N, best first).
func (h *Handler) createRoundOneMatches(tx *sql.Tx, ctx context.Context, bracket *domain.Bracket, topTorrons []*domain.Torro) error {
	order := standardSeedOrder(bracket.Size)
	numMatches := bracket.Size / 2

	for slot := 0; slot < numMatches; slot++ {
		seedA := order[2*slot]
		seedB := order[2*slot+1]

		torroA, okA := torroBySeed(topTorrons, seedA)
		torroB, okB := torroBySeed(topTorrons, seedB)

		if !okA && !okB {
			// Neither seed exists (class far smaller than requested
			// size) - no match at all for this slot.
			continue
		}

		match := &domain.BracketMatch{
			BracketId: bracket.Id,
			Round:     1,
			Slot:      slot,
		}

		switch {
		case okA && okB:
			match.Torro1Id = torroA.Id
			torro2Id := torroB.Id
			match.Torro2Id = &torro2Id
			match.Status = domain.BracketMatchStatusPending
		case okA:
			match.Torro1Id = torroA.Id
			match.Status = domain.BracketMatchStatusCompleted
			winnerId := torroA.Id
			match.WinnerId = &winnerId
		default:
			match.Torro1Id = torroB.Id
			match.Status = domain.BracketMatchStatusCompleted
			winnerId := torroB.Id
			match.WinnerId = &winnerId
		}

		if _, err := h.bracketRepo.CreateMatchTx(tx, ctx, match); err != nil {
			return err
		}
	}

	return nil
}

// bracketAdvance handles POST /bracket/{bracketId}/advance: force round
// advancement regardless of vote completeness. Any match still pending in
// the current round is decided by its current tally, ties (including 0-0
// for an untouched match) broken by lower seed.
//
// Gated by Handler.RequireAdminToken - see its route registration in server.go.
func (h *Handler) bracketAdvance(w http.ResponseWriter, r *http.Request) {
	logger.Info("[Handler - BracketAdvance] Incoming request")

	bracketId := chi.URLParam(r, "bracketId")
	ctx := r.Context()

	tx, err := h.db.Begin()
	if err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't start transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	defer tx.Rollback()

	bracket, err := h.bracketRepo.GetTx(tx, ctx, bracketId)
	if err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't get bracket %s. %v", bracketId, err)
		renderBracketError(w, r, err)
		return
	}

	if bracket.Status == domain.BracketStatusCompleted {
		render.Render(w, r, domain.ErrBadRequest(
			fmt.Errorf("%s: bracket %s is already completed", domain.ValidationError, bracketId)))
		return
	}

	entries, err := h.bracketRepo.ListEntriesTx(tx, ctx, bracket.Id)
	if err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't list entries for bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}
	seedByTorro := seedMap(entries)

	if err := h.resolvePendingMatchesInRound(tx, ctx, bracket, seedByTorro); err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't resolve round %d. %v", bracket.CurrentRound, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	if _, err := h.cascadeAdvance(tx, ctx, bracket); err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't advance bracket %s. %v", bracket.Id, err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	if err := tx.Commit(); err != nil {
		logger.Error("[Handler - BracketAdvance] Couldn't commit transaction. %v", err)
		render.Render(w, r, domain.ErrInternal(err))
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, bracket)
}

// -- helpers --

// standardSeedOrder returns the canonical single-elimination seed order
// for a bracket of the given size (must be a power of two). For size=8
// this returns [1,8,4,5,2,7,3,6]; pairing consecutive entries yields the
// familiar 1v8, 4v5, 2v7, 3v6 first-round matchups. Generalizes to any
// power-of-two size via the standard "reflect and interleave" doubling
// construction.
func standardSeedOrder(size int) []int {
	order := []int{1}
	for len(order) < size {
		n := len(order)
		next := make([]int, 0, n*2)
		for _, x := range order {
			next = append(next, x, 2*n+1-x)
		}
		order = next
	}
	return order
}

func isPowerOfTwo(n int) bool {
	return n > 0 && n&(n-1) == 0
}

// torroBySeed returns the torró at the given 1-indexed seed within a
// seed-ordered slice (as returned by TorroRepo.TopNByClass), and whether
// that seed actually exists (it may not, if the class has fewer active
// torrons than the requested bracket size).
func torroBySeed(seedOrdered []*domain.Torro, seed int) (*domain.Torro, bool) {
	if seed < 1 || seed > len(seedOrdered) {
		return nil, false
	}
	return seedOrdered[seed-1], true
}

func seedMap(entries []*domain.BracketEntry) map[string]int {
	m := make(map[string]int, len(entries))
	for _, e := range entries {
		m[e.TorronId] = e.Seed
	}
	return m
}

func totalVotes(votes map[string]int) int {
	total := 0
	for _, c := range votes {
		total += c
	}
	return total
}

// decideMatchWinner tallies a match's votes and returns the winning
// torró's ID. A tie (including 0-0, which happens when force-advancing an
// untouched match) is broken by the lower original seed number - i.e. the
// torró Phase 1 rated higher advances. This is a judgment call: it favours
// the higher Phase 1 seed over a coin flip, which seems the least
// arbitrary default, but is worth double-checking against intended
// product behaviour.
func decideMatchWinner(match *domain.BracketMatch, votes map[string]int, seedByTorro map[string]int) string {
	if match.IsBye() {
		return match.Torro1Id
	}

	v1 := votes[match.Torro1Id]
	v2 := votes[*match.Torro2Id]

	if v1 > v2 {
		return match.Torro1Id
	}
	if v2 > v1 {
		return *match.Torro2Id
	}

	seed1 := seedByTorro[match.Torro1Id]
	seed2 := seedByTorro[*match.Torro2Id]

	logger.Info(
		"[BracketAdvance] Match %s tied %d-%d, breaking by seed: %d vs %d",
		match.Id, v1, v2, seed1, seed2,
	)

	if seed1 <= seed2 {
		return match.Torro1Id
	}
	return *match.Torro2Id
}

// buildMatchView assembles the display view for a match, fetching torró
// details through getTorro (see torroFetcher).
func buildMatchView(m *domain.BracketMatch, seedByTorro map[string]int, getTorro func(string) (*domain.Torro, error)) (BracketMatchView, error) {
	t1, err := getTorro(m.Torro1Id)
	if err != nil {
		return BracketMatchView{}, err
	}

	view := BracketMatchView{
		Id:     m.Id,
		Round:  m.Round,
		Slot:   m.Slot,
		Status: m.Status,
		IsBye:  m.IsBye(),
		Torro1: BracketTorroView{Id: t1.Id, Name: t1.Name, Image: t1.Image, Seed: seedByTorro[t1.Id]},
	}

	if m.Torro2Id != nil {
		t2, err := getTorro(*m.Torro2Id)
		if err != nil {
			return BracketMatchView{}, err
		}
		view.Torro2 = &BracketTorroView{Id: t2.Id, Name: t2.Name, Image: t2.Image, Seed: seedByTorro[t2.Id]}
	}

	if m.WinnerId != nil {
		view.Decided = true
		switch {
		case *m.WinnerId == view.Torro1.Id:
			view.Torro1Won = true
		case view.Torro2 != nil && *m.WinnerId == view.Torro2.Id:
			view.Torro2Won = true
		}
	}

	return view, nil
}

// torroFetcher returns a request-scoped, memoized torró getter so pages
// that reference the same torró across several matches (e.g. the overview
// page across rounds) don't re-query it repeatedly.
func (h *Handler) torroFetcher(ctx context.Context) func(string) (*domain.Torro, error) {
	cache := make(map[string]*domain.Torro)
	return func(id string) (*domain.Torro, error) {
		if t, ok := cache[id]; ok {
			return t, nil
		}
		t, err := h.torroRepo.Get(ctx, id)
		if err != nil {
			return nil, err
		}
		cache[id] = t
		return t, nil
	}
}

// pickRandomMatch picks one match at random using a cryptographically
// secure source, mirroring PairingRepo.GetRandom's approach. Returns nil
// (no error) if matches is empty.
func pickRandomMatch(matches []*domain.BracketMatch) (*domain.BracketMatch, error) {
	if len(matches) == 0 {
		return nil, nil
	}

	idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(matches))))
	if err != nil {
		return nil, err
	}

	return matches[idx.Int64()], nil
}

// renderBracketError maps a repository error to the appropriate HTTP
// response, following the same domain.ErrXxx convention used throughout
// the rest of the handlers.
func renderBracketError(w http.ResponseWriter, r *http.Request, err error) {
	msg := err.Error()
	switch {
	case strings.Contains(msg, string(domain.NotFoundError)):
		render.Render(w, r, domain.ErrNotFound(err))
	case strings.Contains(msg, string(domain.ValidationError)),
		strings.Contains(msg, string(domain.DuplicateKeyError)),
		strings.Contains(msg, string(domain.ForeignKeyError)):
		render.Render(w, r, domain.ErrBadRequest(err))
	default:
		render.Render(w, r, domain.ErrInternal(err))
	}
}
