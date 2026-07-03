package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"

	"github.com/krtffl/torro/internal/domain"
)

type postgresWrappedStatsRepo struct {
	db *sql.DB
}

// NewWrappedStatsRepo constructs a WrappedStatsRepo backed by Postgres.
// Like PressStatsRepo, every method here is a read-only aggregation over
// existing Phase 1 (Results/Pairings/Torrons) and Phase 2
// (BracketMatches/BracketMatchVotes) tables - no new schema is required
// for the personal "Torrorèndum Wrapped" recap.
func NewWrappedStatsRepo(db *sql.DB) domain.WrappedStatsRepo {
	return &postgresWrappedStatsRepo{
		db: db,
	}
}

// userPairingRow is one row of the per-user-pairing dataset DuelStats
// derives both of its stats from: a pairing the user voted on, their own
// pick, and the crowd's overall split on that same pairing.
type userPairingRow struct {
	userPick   string
	torro1Id   string
	torro2Id   string
	votes1     int
	votes2     int
	totalVotes int
}

// DuelStats fetches every pairing the user voted on that meets the
// minTotalVotes eligibility threshold (join a user's own vote back to the
// full crowd tally for that same pairing), then derives both the most
// contested duel and the most unpopular pick from that single dataset in
// Go, rather than running two separate SQL queries.
func (r *postgresWrappedStatsRepo) DuelStats(ctx context.Context, userId string, minTotalVotes int) (*domain.WrappedDuelStats, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT
            r."Winner" AS user_pick, p."Torro1", p."Torro2",
            COUNT(*) FILTER (WHERE g."Winner" = p."Torro1") AS votes1,
            COUNT(*) FILTER (WHERE g."Winner" = p."Torro2") AS votes2,
            COUNT(*) AS total_votes
        FROM "Results" r
        JOIN "Pairings" p ON r."Pairing" = p."Id"
        JOIN "Results" g ON g."Pairing" = r."Pairing"
        WHERE r."UserId" = $1
        GROUP BY r."Pairing", r."Winner", p."Torro1", p."Torro2"
        HAVING COUNT(*) >= $2`,
		userId,
		minTotalVotes,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var candidates []userPairingRow
	torroIdSet := make(map[string]bool)
	for rows.Next() {
		var row userPairingRow
		if err := rows.Scan(&row.userPick, &row.torro1Id, &row.torro2Id, &row.votes1, &row.votes2, &row.totalVotes); err != nil {
			return nil, handleErrors(err)
		}
		candidates = append(candidates, row)
		torroIdSet[row.torro1Id] = true
		torroIdSet[row.torro2Id] = true
	}
	if err := rows.Err(); err != nil {
		return nil, handleErrors(err)
	}

	stats := &domain.WrappedDuelStats{}
	if len(candidates) == 0 {
		return stats, nil
	}

	ids := make([]string, 0, len(torroIdSet))
	for id := range torroIdSet {
		ids = append(ids, id)
	}
	names, err := r.torroNamesById(ctx, ids)
	if err != nil {
		return nil, err
	}

	// Most contested: minimizes the absolute gap between the two sides'
	// vote counts. Most unpopular: minimizes the user's own pick's share
	// of that pairing's total votes. Both are picked out of the same
	// candidate slice in a single pass.
	var contested *userPairingRow
	var unpopular *userPairingRow
	var unpopularShare float64

	for i := range candidates {
		c := &candidates[i]

		if contested == nil || absInt(c.votes1-c.votes2) < absInt(contested.votes1-contested.votes2) {
			contested = c
		}

		userVotes := c.votes1
		if c.userPick == c.torro2Id {
			userVotes = c.votes2
		}
		share := float64(userVotes) / float64(c.totalVotes)
		if unpopular == nil || share < unpopularShare {
			unpopular = c
			unpopularShare = share
		}
	}

	if contested != nil {
		stats.HasContestedDuel = true
		stats.ContestedDuel = toUserDuelStat(*contested, names)
	}
	if unpopular != nil {
		stats.HasUnpopularPick = true
		stats.UnpopularPick = toUserDuelStat(*unpopular, names)
	}

	return stats, nil
}

// BracketPath summarizes a user's participation in one bracket: how many
// rounds they voted in, how many of those matches are decided, how many
// they called correctly, and (best-effort, independent of whether the
// user voted at all) whether the bracket itself has produced a champion
// and whether the user ever picked that champion in one of their votes.
func (r *postgresWrappedStatsRepo) BracketPath(ctx context.Context, userId string, bracketId string) (*domain.BracketPathStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT m."Round", m."WinnerId", v."TorronId" AS user_pick
        FROM "BracketMatchVotes" v
        JOIN "BracketMatches" m ON m."Id" = v."MatchId"
        WHERE m."BracketId" = $1 AND v."UserId" = $2
        ORDER BY m."Round" ASC`,
		bracketId,
		userId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	stat := &domain.BracketPathStat{}
	rounds := make(map[int]bool)
	userPicks := make(map[string]bool)

	for rows.Next() {
		var round int
		var winnerId sql.NullString
		var userPick string
		if err := rows.Scan(&round, &winnerId, &userPick); err != nil {
			return nil, handleErrors(err)
		}

		stat.HasVoted = true
		rounds[round] = true
		userPicks[userPick] = true

		if winnerId.Valid {
			stat.MatchesDecided++
			if userPick == winnerId.String {
				stat.PicksCorrect++
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, handleErrors(err)
	}
	stat.RoundsVoted = len(rounds)

	if !stat.HasVoted {
		return stat, nil
	}

	// Resolve the champion decision as a small, separate lookup: whether
	// the bracket itself is finished is independent of whether this
	// particular user voted at all, so it can't be derived from the vote
	// rows fetched above.
	var status string
	var championId sql.NullString
	var championName sql.NullString
	row := r.db.QueryRowContext(ctx,
		`
        SELECT b."Status", b."ChampionId", t."Name"
        FROM "Brackets" b
        LEFT JOIN "Torrons" t ON t."Id" = b."ChampionId"
        WHERE b."Id" = $1`,
		bracketId,
	)
	if err := row.Scan(&status, &championId, &championName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return stat, nil
		}
		return nil, handleErrors(err)
	}

	if status == domain.BracketStatusCompleted && championId.Valid {
		stat.HasChampion = true
		stat.ChampionName = championName.String
		stat.MatchedChampion = userPicks[championId.String]
	}

	return stat, nil
}

// toUserDuelStat converts a raw per-user-pairing row plus a name lookup
// into the domain-facing UserDuelStat shape.
func toUserDuelStat(row userPairingRow, names map[string]string) domain.UserDuelStat {
	return domain.UserDuelStat{
		TorroAId:     row.torro1Id,
		TorroAName:   names[row.torro1Id],
		TorroBId:     row.torro2Id,
		TorroBName:   names[row.torro2Id],
		UserPickId:   row.userPick,
		UserPickName: names[row.userPick],
		VotesA:       row.votes1,
		VotesB:       row.votes2,
		TotalVotes:   row.totalVotes,
	}
}

// torroNamesById fetches the Name of an arbitrary set of torró IDs in a
// single round trip, keyed by ID. Unlike postgresPressStatsRepo's
// torroNamesAndImages (which always looks up exactly two IDs for a
// pairing), DuelStats needs an arbitrary-sized batch across every
// candidate pairing, hence the "= ANY($1)" array form instead of a fixed
// "IN ($1, $2)".
func (r *postgresWrappedStatsRepo) torroNamesById(ctx context.Context, ids []string) (map[string]string, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "Name" FROM "Torrons" WHERE "Id" = ANY($1)`,
		pq.Array(ids),
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	names := make(map[string]string, len(ids))
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, handleErrors(err)
		}
		names[id] = name
	}
	if err := rows.Err(); err != nil {
		return nil, handleErrors(err)
	}

	return names, nil
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
