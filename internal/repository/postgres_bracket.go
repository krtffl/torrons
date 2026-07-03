package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresBracketRepo struct {
	db *sql.DB
}

func NewBracketRepo(db *sql.DB) domain.BracketRepo {
	return &postgresBracketRepo{
		db: db,
	}
}

// -- Brackets --

func (r *postgresBracketRepo) Create(ctx context.Context, bracket *domain.Bracket) (*domain.Bracket, error) {
	if bracket.Id == "" {
		bracket.Id = uuid.NewString()
	}
	if bracket.CurrentRound == 0 {
		bracket.CurrentRound = 1
	}
	if bracket.Status == "" {
		bracket.Status = domain.BracketStatusInProgress
	}
	if bracket.CreatedAt == "" {
		bracket.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "Brackets" ("Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING "Id"`,
		bracket.Id,
		bracket.CampaignId,
		bracket.ClassId,
		bracket.Size,
		bracket.CurrentRound,
		bracket.Status,
		bracket.CreatedAt,
	).Scan(&bracket.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return bracket, nil
}

func (r *postgresBracketRepo) Get(ctx context.Context, id string) (*domain.Bracket, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "ChampionId", "CreatedAt", "CompletedAt"
		 FROM "Brackets"
		 WHERE "Id" = $1`,
		id,
	)
	return scanBracket(row)
}

func (r *postgresBracketRepo) GetByCampaignAndClass(ctx context.Context, campaignId string, classId string) (*domain.Bracket, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "ChampionId", "CreatedAt", "CompletedAt"
		 FROM "Brackets"
		 WHERE "CampaignId" = $1 AND "ClassId" = $2`,
		campaignId,
		classId,
	)
	return scanBracket(row)
}

func (r *postgresBracketRepo) GetLatestByClass(ctx context.Context, classId string) (*domain.Bracket, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "ChampionId", "CreatedAt", "CompletedAt"
		 FROM "Brackets"
		 WHERE "ClassId" = $1
		 ORDER BY "CreatedAt" DESC
		 LIMIT 1`,
		classId,
	)
	return scanBracket(row)
}

func (r *postgresBracketRepo) UpdateRound(ctx context.Context, id string, round int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Brackets" SET "CurrentRound" = $2 WHERE "Id" = $1`,
		id,
		round,
	)
	return handleErrors(err)
}

func (r *postgresBracketRepo) Complete(ctx context.Context, id string, championId string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Brackets"
		 SET "Status" = $2, "ChampionId" = $3, "CompletedAt" = $4
		 WHERE "Id" = $1`,
		id,
		domain.BracketStatusCompleted,
		championId,
		time.Now().UTC().Format(time.RFC3339),
	)
	return handleErrors(err)
}

// -- Bracket entries --

func (r *postgresBracketRepo) CreateEntry(ctx context.Context, entry *domain.BracketEntry) (*domain.BracketEntry, error) {
	if entry.Id == "" {
		entry.Id = uuid.NewString()
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "BracketEntries" ("Id", "BracketId", "TorronId", "Seed", "SeedRating")
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING "Id"`,
		entry.Id,
		entry.BracketId,
		entry.TorronId,
		entry.Seed,
		entry.SeedRating,
	).Scan(&entry.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return entry, nil
}

func (r *postgresBracketRepo) ListEntries(ctx context.Context, bracketId string) ([]*domain.BracketEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "BracketId", "TorronId", "Seed", "SeedRating"
		 FROM "BracketEntries"
		 WHERE "BracketId" = $1
		 ORDER BY "Seed" ASC`,
		bracketId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketEntries(rows)
}

// -- Bracket matches --

func (r *postgresBracketRepo) CreateMatch(ctx context.Context, match *domain.BracketMatch) (*domain.BracketMatch, error) {
	if match.Id == "" {
		match.Id = uuid.NewString()
	}
	if match.Status == "" {
		match.Status = domain.BracketMatchStatusPending
	}
	if match.CreatedAt == "" {
		match.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "BracketMatches" ("Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING "Id"`,
		match.Id,
		match.BracketId,
		match.Round,
		match.Slot,
		match.Torro1Id,
		match.Torro2Id,
		match.WinnerId,
		match.Status,
		match.CreatedAt,
	).Scan(&match.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return match, nil
}

func (r *postgresBracketRepo) GetMatch(ctx context.Context, id string) (*domain.BracketMatch, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt"
		 FROM "BracketMatches"
		 WHERE "Id" = $1`,
		id,
	)
	return scanBracketMatch(row)
}

func (r *postgresBracketRepo) ListMatchesByRound(ctx context.Context, bracketId string, round int) ([]*domain.BracketMatch, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt"
		 FROM "BracketMatches"
		 WHERE "BracketId" = $1 AND "Round" = $2
		 ORDER BY "Slot" ASC`,
		bracketId,
		round,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketMatches(rows)
}

func (r *postgresBracketRepo) ListMatches(ctx context.Context, bracketId string) ([]*domain.BracketMatch, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt"
		 FROM "BracketMatches"
		 WHERE "BracketId" = $1
		 ORDER BY "Round" ASC, "Slot" ASC`,
		bracketId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketMatches(rows)
}

func (r *postgresBracketRepo) ListOpenMatchesForUser(ctx context.Context, bracketId string, round int, userId string) ([]*domain.BracketMatch, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT m."Id", m."BracketId", m."Round", m."Slot", m."Torro1Id", m."Torro2Id", m."WinnerId", m."Status", m."CreatedAt"
		 FROM "BracketMatches" m
		 WHERE m."BracketId" = $1
		   AND m."Round" = $2
		   AND m."Status" = $3
		   AND NOT EXISTS (
		       SELECT 1 FROM "BracketMatchVotes" v
		       WHERE v."MatchId" = m."Id" AND v."UserId" = $4
		   )
		 ORDER BY m."Slot" ASC`,
		bracketId,
		round,
		domain.BracketMatchStatusPending,
		userId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketMatches(rows)
}

func (r *postgresBracketRepo) SetMatchWinner(ctx context.Context, matchId string, winnerId string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "BracketMatches"
		 SET "WinnerId" = $2, "Status" = $3
		 WHERE "Id" = $1`,
		matchId,
		winnerId,
		domain.BracketMatchStatusCompleted,
	)
	return handleErrors(err)
}

// -- Votes --

func (r *postgresBracketRepo) CreateVote(ctx context.Context, vote *domain.BracketMatchVote) (*domain.BracketMatchVote, error) {
	if vote.Id == "" {
		vote.Id = uuid.NewString()
	}
	if vote.CreatedAt == "" {
		vote.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "BracketMatchVotes" ("Id", "MatchId", "UserId", "TorronId", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING "Id"`,
		vote.Id,
		vote.MatchId,
		vote.UserId,
		vote.TorronId,
		vote.CreatedAt,
	).Scan(&vote.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return vote, nil
}

func (r *postgresBracketRepo) HasVoted(ctx context.Context, matchId string, userId string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
		     SELECT 1 FROM "BracketMatchVotes" WHERE "MatchId" = $1 AND "UserId" = $2
		 )`,
		matchId,
		userId,
	).Scan(&exists)
	if err != nil {
		return false, handleErrors(err)
	}

	return exists, nil
}

func (r *postgresBracketRepo) CountVotesByTorron(ctx context.Context, matchId string) (map[string]int, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "TorronId", COUNT(*)
		 FROM "BracketMatchVotes"
		 WHERE "MatchId" = $1
		 GROUP BY "TorronId"`,
		matchId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanVoteCounts(rows)
}

// -- Transaction methods --

func (r *postgresBracketRepo) CreateTx(tx *sql.Tx, ctx context.Context, bracket *domain.Bracket) (*domain.Bracket, error) {
	if bracket.Id == "" {
		bracket.Id = uuid.NewString()
	}
	if bracket.CurrentRound == 0 {
		bracket.CurrentRound = 1
	}
	if bracket.Status == "" {
		bracket.Status = domain.BracketStatusInProgress
	}
	if bracket.CreatedAt == "" {
		bracket.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := tx.QueryRowContext(ctx,
		`INSERT INTO "Brackets" ("Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING "Id"`,
		bracket.Id,
		bracket.CampaignId,
		bracket.ClassId,
		bracket.Size,
		bracket.CurrentRound,
		bracket.Status,
		bracket.CreatedAt,
	).Scan(&bracket.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return bracket, nil
}

func (r *postgresBracketRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.Bracket, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT "Id", "CampaignId", "ClassId", "Size", "CurrentRound", "Status", "ChampionId", "CreatedAt", "CompletedAt"
		 FROM "Brackets"
		 WHERE "Id" = $1`,
		id,
	)
	return scanBracket(row)
}

func (r *postgresBracketRepo) UpdateRoundTx(tx *sql.Tx, ctx context.Context, id string, round int) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE "Brackets" SET "CurrentRound" = $2 WHERE "Id" = $1`,
		id,
		round,
	)
	return handleErrors(err)
}

func (r *postgresBracketRepo) CompleteTx(tx *sql.Tx, ctx context.Context, id string, championId string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE "Brackets"
		 SET "Status" = $2, "ChampionId" = $3, "CompletedAt" = $4
		 WHERE "Id" = $1`,
		id,
		domain.BracketStatusCompleted,
		championId,
		time.Now().UTC().Format(time.RFC3339),
	)
	return handleErrors(err)
}

func (r *postgresBracketRepo) CreateEntryTx(tx *sql.Tx, ctx context.Context, entry *domain.BracketEntry) (*domain.BracketEntry, error) {
	if entry.Id == "" {
		entry.Id = uuid.NewString()
	}

	err := tx.QueryRowContext(ctx,
		`INSERT INTO "BracketEntries" ("Id", "BracketId", "TorronId", "Seed", "SeedRating")
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING "Id"`,
		entry.Id,
		entry.BracketId,
		entry.TorronId,
		entry.Seed,
		entry.SeedRating,
	).Scan(&entry.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return entry, nil
}

func (r *postgresBracketRepo) ListEntriesTx(tx *sql.Tx, ctx context.Context, bracketId string) ([]*domain.BracketEntry, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT "Id", "BracketId", "TorronId", "Seed", "SeedRating"
		 FROM "BracketEntries"
		 WHERE "BracketId" = $1
		 ORDER BY "Seed" ASC`,
		bracketId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketEntries(rows)
}

func (r *postgresBracketRepo) CreateMatchTx(tx *sql.Tx, ctx context.Context, match *domain.BracketMatch) (*domain.BracketMatch, error) {
	if match.Id == "" {
		match.Id = uuid.NewString()
	}
	if match.Status == "" {
		match.Status = domain.BracketMatchStatusPending
	}
	if match.CreatedAt == "" {
		match.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := tx.QueryRowContext(ctx,
		`INSERT INTO "BracketMatches" ("Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		 RETURNING "Id"`,
		match.Id,
		match.BracketId,
		match.Round,
		match.Slot,
		match.Torro1Id,
		match.Torro2Id,
		match.WinnerId,
		match.Status,
		match.CreatedAt,
	).Scan(&match.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return match, nil
}

func (r *postgresBracketRepo) GetMatchTx(tx *sql.Tx, ctx context.Context, id string) (*domain.BracketMatch, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT "Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt"
		 FROM "BracketMatches"
		 WHERE "Id" = $1`,
		id,
	)
	return scanBracketMatch(row)
}

func (r *postgresBracketRepo) ListMatchesByRoundTx(tx *sql.Tx, ctx context.Context, bracketId string, round int) ([]*domain.BracketMatch, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT "Id", "BracketId", "Round", "Slot", "Torro1Id", "Torro2Id", "WinnerId", "Status", "CreatedAt"
		 FROM "BracketMatches"
		 WHERE "BracketId" = $1 AND "Round" = $2
		 ORDER BY "Slot" ASC`,
		bracketId,
		round,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanBracketMatches(rows)
}

func (r *postgresBracketRepo) SetMatchWinnerTx(tx *sql.Tx, ctx context.Context, matchId string, winnerId string) error {
	_, err := tx.ExecContext(ctx,
		`UPDATE "BracketMatches"
		 SET "WinnerId" = $2, "Status" = $3
		 WHERE "Id" = $1`,
		matchId,
		winnerId,
		domain.BracketMatchStatusCompleted,
	)
	return handleErrors(err)
}

func (r *postgresBracketRepo) CreateVoteTx(tx *sql.Tx, ctx context.Context, vote *domain.BracketMatchVote) (*domain.BracketMatchVote, error) {
	if vote.Id == "" {
		vote.Id = uuid.NewString()
	}
	if vote.CreatedAt == "" {
		vote.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := tx.QueryRowContext(ctx,
		`INSERT INTO "BracketMatchVotes" ("Id", "MatchId", "UserId", "TorronId", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING "Id"`,
		vote.Id,
		vote.MatchId,
		vote.UserId,
		vote.TorronId,
		vote.CreatedAt,
	).Scan(&vote.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return vote, nil
}

func (r *postgresBracketRepo) CountVotesByTorronTx(tx *sql.Tx, ctx context.Context, matchId string) (map[string]int, error) {
	rows, err := tx.QueryContext(ctx,
		`SELECT "TorronId", COUNT(*)
		 FROM "BracketMatchVotes"
		 WHERE "MatchId" = $1
		 GROUP BY "TorronId"`,
		matchId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanVoteCounts(rows)
}

// -- scanning helpers --

// row is satisfied by both *sql.Row and *sql.Rows (the subset used here).
type row interface {
	Scan(dest ...interface{}) error
}

func scanBracket(row row) (*domain.Bracket, error) {
	bracket := &domain.Bracket{}
	var championId sql.NullString
	var completedAt sql.NullString

	err := row.Scan(
		&bracket.Id,
		&bracket.CampaignId,
		&bracket.ClassId,
		&bracket.Size,
		&bracket.CurrentRound,
		&bracket.Status,
		&championId,
		&bracket.CreatedAt,
		&completedAt,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	if championId.Valid {
		bracket.ChampionId = &championId.String
	}
	if completedAt.Valid {
		bracket.CompletedAt = &completedAt.String
	}

	return bracket, nil
}

func scanBracketEntries(rows *sql.Rows) ([]*domain.BracketEntry, error) {
	var entries []*domain.BracketEntry
	for rows.Next() {
		entry := &domain.BracketEntry{}
		if err := rows.Scan(
			&entry.Id,
			&entry.BracketId,
			&entry.TorronId,
			&entry.Seed,
			&entry.SeedRating,
		); err != nil {
			return nil, handleErrors(err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func scanBracketMatch(row row) (*domain.BracketMatch, error) {
	match := &domain.BracketMatch{}
	var torro2Id sql.NullString
	var winnerId sql.NullString

	err := row.Scan(
		&match.Id,
		&match.BracketId,
		&match.Round,
		&match.Slot,
		&match.Torro1Id,
		&torro2Id,
		&winnerId,
		&match.Status,
		&match.CreatedAt,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	if torro2Id.Valid {
		match.Torro2Id = &torro2Id.String
	}
	if winnerId.Valid {
		match.WinnerId = &winnerId.String
	}

	return match, nil
}

func scanBracketMatches(rows *sql.Rows) ([]*domain.BracketMatch, error) {
	var matches []*domain.BracketMatch
	for rows.Next() {
		match, err := scanBracketMatch(rows)
		if err != nil {
			return nil, err
		}
		matches = append(matches, match)
	}

	return matches, nil
}

func scanVoteCounts(rows *sql.Rows) (map[string]int, error) {
	counts := make(map[string]int)
	for rows.Next() {
		var torronId string
		var count int
		if err := rows.Scan(&torronId, &count); err != nil {
			return nil, handleErrors(err)
		}
		counts[torronId] = count
	}

	return counts, nil
}
