package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/krtffl/torro/internal/domain"
)

type postgresPersonaRepo struct {
	db *sql.DB
}

// NewPersonaRepo constructs a PersonaRepo backed by Postgres. Like
// WrappedStatsRepo/PressStatsRepo, every method here is a read-only
// aggregation over existing Users/Results/Pairings/Torrons/Classes tables -
// no new schema is required for the "torró personality reveal".
func NewPersonaRepo(db *sql.DB) domain.PersonaRepo {
	return &postgresPersonaRepo{
		db: db,
	}
}

// Stats computes one user's PersonaStats - see the doc comment on
// domain.PersonaRepo.Stats for the gating precondition callers must
// already have checked.
func (r *postgresPersonaRepo) Stats(ctx context.Context, userId string, minVotes int) (*domain.PersonaStats, error) {
	votes, totalVotes, err := r.userClassVotes(ctx, userId)
	if err != nil {
		return nil, err
	}

	stats := &domain.PersonaStats{
		HasEnoughVotes: true,
		TotalVotes:     totalVotes,
	}

	// arenaFilter is nil ("no filter", every arena) for the tie case, and a
	// pointer to the winning arena's class id otherwise - see topTorro's
	// doc comment for why a single query serves both shapes.
	var arenaFilter *string

	topClassId, hasFavorite := domain.TopClassIdFromVotes(votes)
	stats.HasClearFavorite = hasFavorite

	if hasFavorite {
		stats.TopClassId = topClassId

		className, err := r.className(ctx, topClassId)
		if err != nil {
			return nil, err
		}
		stats.TopClassName = className

		cohort, err := r.cohortTopClassIds(ctx, minVotes)
		if err != nil {
			return nil, err
		}
		stats.Percentile = percentileOf(topClassId, cohort)

		arenaFilter = &topClassId
	}

	torroId, torroName, votesCast, wins, losses, err := r.topTorro(ctx, userId, arenaFilter)
	if err != nil {
		return nil, err
	}
	stats.TopTorroId = torroId
	stats.TopTorroName = torroName
	stats.TopTorroVotesCast = votesCast
	stats.TopTorroWins = wins
	stats.TopTorroLosses = losses

	return stats, nil
}

// userClassVotes fetches one user's ClassVotes (parsed into a plain
// map[string]int) and their all-time VoteCount in a single round trip.
func (r *postgresPersonaRepo) userClassVotes(ctx context.Context, userId string) (map[string]int, int, error) {
	var raw json.RawMessage
	var voteCount int

	err := r.db.QueryRowContext(ctx,
		`SELECT "ClassVotes", "VoteCount" FROM "Users" WHERE "Id" = $1`,
		userId,
	).Scan(&raw, &voteCount)
	if err != nil {
		return nil, 0, handleErrors(err)
	}

	votes, err := parseClassVotes(raw)
	if err != nil {
		return nil, 0, err
	}

	return votes, voteCount, nil
}

// parseClassVotes unmarshals a Users.ClassVotes JSONB value
// ({"1": 15, "2": 30, ...}) into a plain map, treating a NULL/empty value
// as "no votes in any arena yet" rather than an error.
func parseClassVotes(raw json.RawMessage) (map[string]int, error) {
	votes := map[string]int{}
	if len(raw) == 0 {
		return votes, nil
	}
	if err := json.Unmarshal(raw, &votes); err != nil {
		return nil, fmt.Errorf("persona: parse ClassVotes: %w", err)
	}
	return votes, nil
}

// personaCohortCacheTTL bounds how often cohortTopClassIds actually queries
// the database. A cache miss scans every user with VoteCount >= minVotes and
// parses their ClassVotes JSONB in Go - on every /reveal request, for a
// cohort-wide distribution that only shifts as users cross the vote
// threshold. A short in-process TTL collapses that to at most one refresh
// per window, the same treatment already given to the /premsa stats
// (press_handler.go's pressStatsCacheTTL).
const personaCohortCacheTTL = 5 * time.Minute

// personaCohortCache memoizes cohortTopClassIds's result, keyed by the
// minVotes threshold it was computed for (in practice always the same fixed
// value - see getMinVotesForClass("5") in reveal_handler.go - but keying on
// it defensively costs nothing and stays correct if that ever changes). mu
// guards the cached value; refresh serializes recomputation so a burst of
// requests arriving at expiry collapses into a single DB pass instead of a
// stampede.
var personaCohortCache struct {
	mu       sync.RWMutex
	minVotes int
	topIds   []string
	expiry   time.Time
	hasData  bool

	refresh sync.Mutex
}

// cohortTopClassIds returns the TopClassId of every user who has cleared
// minVotes AND has a single clear-favorite arena themselves (ties are
// excluded from the cohort, same rule as the current user's own
// HasClearFavorite check) - exactly the denominator/numerator inputs
// percentileOf needs. Served from a short-TTL in-process cache (see
// personaCohortCacheTTL) since a miss scans every qualifying user.
func (r *postgresPersonaRepo) cohortTopClassIds(ctx context.Context, minVotes int) ([]string, error) {
	// Fast path: a fresh cached value for this exact threshold serves the
	// vast majority of requests under a single read lock.
	personaCohortCache.mu.RLock()
	if personaCohortCache.hasData && personaCohortCache.minVotes == minVotes && time.Now().Before(personaCohortCache.expiry) {
		ids := personaCohortCache.topIds
		personaCohortCache.mu.RUnlock()
		return ids, nil
	}
	personaCohortCache.mu.RUnlock()

	// Stale, cold, or a different threshold: serialize the refresh.
	personaCohortCache.refresh.Lock()
	defer personaCohortCache.refresh.Unlock()

	// Re-check: another goroutine may have refreshed while we waited.
	personaCohortCache.mu.RLock()
	if personaCohortCache.hasData && personaCohortCache.minVotes == minVotes && time.Now().Before(personaCohortCache.expiry) {
		ids := personaCohortCache.topIds
		personaCohortCache.mu.RUnlock()
		return ids, nil
	}
	personaCohortCache.mu.RUnlock()

	topIds, err := r.computeCohortTopClassIds(ctx, minVotes)
	if err != nil {
		// Fall back to the last good value on a transient failure, but only
		// if it was computed for this SAME threshold - a cache primed for a
		// different minVotes would silently misreport the percentile.
		personaCohortCache.mu.RLock()
		defer personaCohortCache.mu.RUnlock()
		if personaCohortCache.hasData && personaCohortCache.minVotes == minVotes {
			return personaCohortCache.topIds, nil
		}
		return nil, err
	}

	personaCohortCache.mu.Lock()
	personaCohortCache.minVotes = minVotes
	personaCohortCache.topIds = topIds
	personaCohortCache.expiry = time.Now().Add(personaCohortCacheTTL)
	personaCohortCache.hasData = true
	personaCohortCache.mu.Unlock()

	return topIds, nil
}

// computeCohortTopClassIds runs the actual aggregation cohortTopClassIds
// serves from cache.
func (r *postgresPersonaRepo) computeCohortTopClassIds(ctx context.Context, minVotes int) ([]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "ClassVotes" FROM "Users" WHERE "VoteCount" >= $1`,
		minVotes,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var topClassIds []string
	for rows.Next() {
		var raw json.RawMessage
		if err := rows.Scan(&raw); err != nil {
			return nil, handleErrors(err)
		}

		votes, err := parseClassVotes(raw)
		if err != nil {
			return nil, err
		}

		if topId, ok := domain.TopClassIdFromVotes(votes); ok {
			topClassIds = append(topClassIds, topId)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, handleErrors(err)
	}

	return topClassIds, nil
}

// percentileOf returns round(100 * matches / len(cohort)), where matches is
// how many cohort entries equal topClassId. Guarded against an empty
// cohort defensively, even though PersonaRepo.Stats's doc comment
// guarantees this can't happen in practice (the current user's own
// TopClassId is always itself one entry of cohort).
func percentileOf(topClassId string, cohort []string) int {
	if len(cohort) == 0 {
		return 0
	}

	matches := 0
	for _, id := range cohort {
		if id == topClassId {
			matches++
		}
	}

	return int(math.Round(100 * float64(matches) / float64(len(cohort))))
}

// className looks up a single Classes.Name by id.
func (r *postgresPersonaRepo) className(ctx context.Context, classId string) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx,
		`SELECT "Name" FROM "Classes" WHERE "Id" = $1`,
		classId,
	).Scan(&name)
	if err != nil {
		return "", handleErrors(err)
	}
	return name, nil
}

// topTorro finds, among the user's Results within a single arena (classId
// non-nil) or across every arena (classId nil), the torró they voted for
// (won) most often, plus how many votes they cast in that same scope and
// that torró's own win/loss record within it. classId is passed straight
// through as a query arg (database/sql's default converter turns a nil
// *string into SQL NULL, the same pattern already used for
// domain.Result.UserId in postgres_result.go) so a single query text covers
// both the single-arena and every-arena shapes via `$2::text IS NULL OR
// p."Class" = $2` rather than duplicating this method for each shape.
func (r *postgresPersonaRepo) topTorro(ctx context.Context, userId string, classId *string) (torroId, name string, votesCast, wins, losses int, err error) {
	err = r.db.QueryRowContext(ctx,
		`
        SELECT r."Winner"
        FROM "Results" r
        JOIN "Pairings" p ON p."Id" = r."Pairing"
        WHERE r."UserId" = $1 AND ($2::text IS NULL OR p."Class" = $2)
        GROUP BY r."Winner"
        ORDER BY COUNT(*) DESC, r."Winner" ASC
        LIMIT 1`,
		userId, classId,
	).Scan(&torroId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// No votes in scope yet - a legitimate empty state (see
			// domain.PersonaStats's doc comment), not an error. Shouldn't
			// happen in practice given this method is only ever called
			// once the user has cleared the overall reveal-unlock
			// threshold, but handled defensively regardless.
			return "", "", 0, 0, 0, nil
		}
		return "", "", 0, 0, 0, handleErrors(err)
	}

	if err = r.db.QueryRowContext(ctx,
		`
        SELECT COUNT(*)
        FROM "Results" r
        JOIN "Pairings" p ON p."Id" = r."Pairing"
        WHERE r."UserId" = $1 AND ($2::text IS NULL OR p."Class" = $2)`,
		userId, classId,
	).Scan(&votesCast); err != nil {
		return "", "", 0, 0, 0, handleErrors(err)
	}

	if err = r.db.QueryRowContext(ctx,
		`
        SELECT
            COUNT(*) FILTER (WHERE r."Winner" = $3),
            COUNT(*) FILTER (WHERE r."Winner" != $3)
        FROM "Results" r
        JOIN "Pairings" p ON p."Id" = r."Pairing"
        WHERE r."UserId" = $1 AND ($2::text IS NULL OR p."Class" = $2)
          AND (p."Torro1" = $3 OR p."Torro2" = $3)`,
		userId, classId, torroId,
	).Scan(&wins, &losses); err != nil {
		return "", "", 0, 0, 0, handleErrors(err)
	}

	name, err = r.torroName(ctx, torroId)
	if err != nil {
		return "", "", 0, 0, 0, err
	}

	return torroId, name, votesCast, wins, losses, nil
}

// torroName looks up a single Torrons.Name by id.
func (r *postgresPersonaRepo) torroName(ctx context.Context, torroId string) (string, error) {
	var name string
	err := r.db.QueryRowContext(ctx,
		`SELECT "Name" FROM "Torrons" WHERE "Id" = $1`,
		torroId,
	).Scan(&name)
	if err != nil {
		return "", handleErrors(err)
	}
	return name, nil
}
