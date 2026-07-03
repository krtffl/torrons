package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/krtffl/torro/internal/domain"
)

type postgresPressStatsRepo struct {
	db *sql.DB
}

// NewPressStatsRepo constructs a PressStatsRepo backed by Postgres. Every
// method is a read-only aggregation over the existing Results/Pairings/
// Torrons tables - no new schema is required for the /premsa press page.
func NewPressStatsRepo(db *sql.DB) domain.PressStatsRepo {
	return &postgresPressStatsRepo{
		db: db,
	}
}

// MostVotedTorro returns the torró that has been chosen as the winner most
// often across every Result ever recorded.
func (r *postgresPressStatsRepo) MostVotedTorro(ctx context.Context) (*domain.TorroStat, error) {
	row := r.db.QueryRowContext(ctx,
		`
        SELECT t."Id", t."Name", t."Image", COUNT(*) AS "Votes"
        FROM "Results" res
        JOIN "Torrons" t ON t."Id" = res."Winner"
        GROUP BY t."Id", t."Name", t."Image"
        ORDER BY "Votes" DESC
        LIMIT 1`,
	)

	stat := &domain.TorroStat{}
	var votes int64
	if err := row.Scan(&stat.TorroId, &stat.Name, &stat.Image, &votes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, handleErrors(err)
	}
	stat.Value = float64(votes)

	return stat, nil
}

// BiggestRiser returns the torró whose ELO rating rose the most (summed net
// change across every Result it took part in) within the last windowDays
// days. Every Results row records a before/after rating for both sides of
// its Pairing, so this unions the Torro1-side and Torro2-side deltas to get
// one per-torró time series regardless of which side of the pairing a given
// Result put it on.
func (r *postgresPressStatsRepo) BiggestRiser(ctx context.Context, windowDays int) (*domain.TorroStat, error) {
	row := r.db.QueryRowContext(ctx,
		`
        SELECT t."Id", t."Name", t."Image", SUM(deltas."Delta") AS "NetChange"
        FROM (
            SELECT p."Torro1" AS "TorroId",
                   (res."Torro1RatingAfter" - res."Torro1RatingBefore") AS "Delta"
            FROM "Results" res
            JOIN "Pairings" p ON res."Pairing" = p."Id"
            WHERE res."Timestamp" > NOW() - make_interval(days => $1)
            UNION ALL
            SELECT p."Torro2" AS "TorroId",
                   (res."Torro2RatingAfter" - res."Torro2RatingBefore") AS "Delta"
            FROM "Results" res
            JOIN "Pairings" p ON res."Pairing" = p."Id"
            WHERE res."Timestamp" > NOW() - make_interval(days => $1)
        ) deltas
        JOIN "Torrons" t ON t."Id" = deltas."TorroId"
        GROUP BY t."Id", t."Name", t."Image"
        ORDER BY "NetChange" DESC
        LIMIT 1`,
		windowDays,
	)

	stat := &domain.TorroStat{}
	if err := row.Scan(&stat.TorroId, &stat.Name, &stat.Image, &stat.Value); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, handleErrors(err)
	}

	return stat, nil
}

// ClosestDuel returns the pairing (two torrons) with the tightest overall
// vote-count gap, among pairings that have accumulated at least
// minTotalVotes votes in total. Each distinct 2-torró matchup is exactly
// one Pairings row (pairings are de-duplicated at creation time), so
// grouping Results by Pairing reliably identifies one duel.
func (r *postgresPressStatsRepo) ClosestDuel(ctx context.Context, minTotalVotes int) (*domain.ClosestDuel, error) {
	row := r.db.QueryRowContext(ctx,
		`
        SELECT p."Torro1", p."Torro2",
               COUNT(*) FILTER (WHERE res."Winner" = p."Torro1") AS "Votes1",
               COUNT(*) FILTER (WHERE res."Winner" = p."Torro2") AS "Votes2",
               COUNT(*) AS "TotalVotes"
        FROM "Results" res
        JOIN "Pairings" p ON res."Pairing" = p."Id"
        GROUP BY res."Pairing", p."Torro1", p."Torro2"
        HAVING COUNT(*) >= $1
        ORDER BY ABS(
                     COUNT(*) FILTER (WHERE res."Winner" = p."Torro1") -
                     COUNT(*) FILTER (WHERE res."Winner" = p."Torro2")
                 ) ASC,
                 COUNT(*) DESC
        LIMIT 1`,
		minTotalVotes,
	)

	var torro1Id, torro2Id string
	var votes1, votes2, totalVotes int
	if err := row.Scan(&torro1Id, &torro2Id, &votes1, &votes2, &totalVotes); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, handleErrors(err)
	}

	info, err := r.torroNamesAndImages(ctx, torro1Id, torro2Id)
	if err != nil {
		return nil, err
	}

	torroA := info[torro1Id]
	torroA.TorroId = torro1Id
	torroA.Value = float64(votes1)

	torroB := info[torro2Id]
	torroB.TorroId = torro2Id
	torroB.Value = float64(votes2)

	return &domain.ClosestDuel{
		TorroA:     torroA,
		TorroB:     torroB,
		TotalVotes: totalVotes,
	}, nil
}

// torroNamesAndImages fetches the Name/Image of a small set of torró IDs in
// a single round trip, keyed by ID. Used by ClosestDuel to resolve display
// info for both sides of a pairing.
func (r *postgresPressStatsRepo) torroNamesAndImages(ctx context.Context, torro1Id, torro2Id string) (map[string]domain.TorroStat, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "Name", "Image" FROM "Torrons" WHERE "Id" IN ($1, $2)`,
		torro1Id,
		torro2Id,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	info := make(map[string]domain.TorroStat, 2)
	for rows.Next() {
		var stat domain.TorroStat
		if err := rows.Scan(&stat.TorroId, &stat.Name, &stat.Image); err != nil {
			return nil, handleErrors(err)
		}
		info[stat.TorroId] = stat
	}
	if err := rows.Err(); err != nil {
		return nil, handleErrors(err)
	}

	return info, nil
}
