package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresUserEloSnapshotRepo struct {
	db *sql.DB
}

func NewUserEloSnapshotRepo(db *sql.DB) domain.UserEloSnapshotRepo {
	return &postgresUserEloSnapshotRepo{
		db: db,
	}
}

func (r *postgresUserEloSnapshotRepo) Get(ctx context.Context, userId string, torronId string) (*domain.UserEloSnapshot, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "UserId", "TorronId", "Rating", "VoteCount", "LastUpdated"
		 FROM "UserEloSnapshots"
		 WHERE "UserId" = $1 AND "TorronId" = $2`,
		userId,
		torronId,
	)

	snapshot := &domain.UserEloSnapshot{}
	err := row.Scan(
		&snapshot.Id,
		&snapshot.UserId,
		&snapshot.TorronId,
		&snapshot.Rating,
		&snapshot.VoteCount,
		&snapshot.LastUpdated,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return snapshot, nil
}

func (r *postgresUserEloSnapshotRepo) Create(ctx context.Context, snapshot *domain.UserEloSnapshot) (*domain.UserEloSnapshot, error) {
	if snapshot.Id == "" {
		snapshot.Id = uuid.NewString()
	}

	if snapshot.LastUpdated == "" {
		snapshot.LastUpdated = time.Now().UTC().Format(time.RFC3339)
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "UserEloSnapshots" ("Id", "UserId", "TorronId", "Rating", "VoteCount", "LastUpdated")
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING "Id"`,
		snapshot.Id,
		snapshot.UserId,
		snapshot.TorronId,
		snapshot.Rating,
		snapshot.VoteCount,
		snapshot.LastUpdated,
	).Scan(&snapshot.Id)

	if err != nil {
		return nil, handleErrors(err)
	}

	return snapshot, nil
}

func (r *postgresUserEloSnapshotRepo) Update(ctx context.Context, snapshot *domain.UserEloSnapshot) (*domain.UserEloSnapshot, error) {
	snapshot.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	_, err := r.db.ExecContext(ctx,
		`UPDATE "UserEloSnapshots"
		 SET "Rating" = $3, "VoteCount" = $4, "LastUpdated" = $5
		 WHERE "UserId" = $1 AND "TorronId" = $2`,
		snapshot.UserId,
		snapshot.TorronId,
		snapshot.Rating,
		snapshot.VoteCount,
		snapshot.LastUpdated,
	)

	if err != nil {
		return nil, handleErrors(err)
	}

	return snapshot, nil
}

func (r *postgresUserEloSnapshotRepo) GetOrCreate(ctx context.Context, userId string, torronId string) (*domain.UserEloSnapshot, error) {
	// Try to get existing snapshot
	snapshot, err := r.Get(ctx, userId, torronId)
	if err == nil {
		return snapshot, nil
	}

	// If not found, get global rating as baseline and create new snapshot
	var globalRating float64
	err = r.db.QueryRowContext(ctx,
		`SELECT "Rating" FROM "Torrons" WHERE "Id" = $1`,
		torronId,
	).Scan(&globalRating)

	if err != nil {
		return nil, handleErrors(err)
	}

	// Create new snapshot with global rating as baseline
	newSnapshot := &domain.UserEloSnapshot{
		UserId:    userId,
		TorronId:  torronId,
		Rating:    globalRating,
		VoteCount: 0,
	}

	return r.Create(ctx, newSnapshot)
}

func (r *postgresUserEloSnapshotRepo) GetUserLeaderboard(ctx context.Context, userId string, classId string) ([]*domain.UserLeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			ues."TorronId",
			t."Name",
			t."Image",
			ues."Rating",
			ues."VoteCount",
			RANK() OVER (ORDER BY ues."Rating" DESC) as rank
		 FROM "UserEloSnapshots" ues
		 INNER JOIN "Torrons" t ON ues."TorronId" = t."Id"
		 WHERE ues."UserId" = $1
		   AND t."Class" = $2
		   AND t."Discontinued" = false
		 ORDER BY ues."Rating" DESC`,
		userId,
		classId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var entries []*domain.UserLeaderboardEntry
	for rows.Next() {
		entry := &domain.UserLeaderboardEntry{}
		err := rows.Scan(
			&entry.TorronId,
			&entry.TorronName,
			&entry.TorronImage,
			&entry.Rating,
			&entry.VoteCount,
			&entry.Rank,
		)
		if err != nil {
			return nil, handleErrors(err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *postgresUserEloSnapshotRepo) GetUserGlobalLeaderboard(ctx context.Context, userId string) ([]*domain.UserLeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			ues."TorronId",
			t."Name",
			t."Image",
			ues."Rating",
			ues."VoteCount",
			RANK() OVER (ORDER BY ues."Rating" DESC) as rank
		 FROM "UserEloSnapshots" ues
		 INNER JOIN "Torrons" t ON ues."TorronId" = t."Id"
		 WHERE ues."UserId" = $1
		   AND t."Discontinued" = false
		 ORDER BY ues."Rating" DESC
		 LIMIT 100`,
		userId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var entries []*domain.UserLeaderboardEntry
	for rows.Next() {
		entry := &domain.UserLeaderboardEntry{}
		err := rows.Scan(
			&entry.TorronId,
			&entry.TorronName,
			&entry.TorronImage,
			&entry.Rating,
			&entry.VoteCount,
			&entry.Rank,
		)
		if err != nil {
			return nil, handleErrors(err)
		}
		entries = append(entries, entry)
	}

	return entries, nil
}

func (r *postgresUserEloSnapshotRepo) ListByUser(ctx context.Context, userId string) ([]*domain.UserEloSnapshot, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "UserId", "TorronId", "Rating", "VoteCount", "LastUpdated"
		 FROM "UserEloSnapshots"
		 WHERE "UserId" = $1
		 ORDER BY "Rating" DESC`,
		userId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var snapshots []*domain.UserEloSnapshot
	for rows.Next() {
		snapshot := &domain.UserEloSnapshot{}
		err := rows.Scan(
			&snapshot.Id,
			&snapshot.UserId,
			&snapshot.TorronId,
			&snapshot.Rating,
			&snapshot.VoteCount,
			&snapshot.LastUpdated,
		)
		if err != nil {
			return nil, handleErrors(err)
		}
		snapshots = append(snapshots, snapshot)
	}

	return snapshots, nil
}

// Transaction methods

func (r *postgresUserEloSnapshotRepo) GetTx(tx *sql.Tx, ctx context.Context, userId string, torronId string) (*domain.UserEloSnapshot, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT "Id", "UserId", "TorronId", "Rating", "VoteCount", "LastUpdated"
		 FROM "UserEloSnapshots"
		 WHERE "UserId" = $1 AND "TorronId" = $2`,
		userId,
		torronId,
	)

	snapshot := &domain.UserEloSnapshot{}
	err := row.Scan(
		&snapshot.Id,
		&snapshot.UserId,
		&snapshot.TorronId,
		&snapshot.Rating,
		&snapshot.VoteCount,
		&snapshot.LastUpdated,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return snapshot, nil
}

func (r *postgresUserEloSnapshotRepo) GetOrCreateTx(tx *sql.Tx, ctx context.Context, userId string, torronId string) (*domain.UserEloSnapshot, error) {
	// Try to get existing snapshot
	snapshot, err := r.GetTx(tx, ctx, userId, torronId)
	if err == nil {
		return snapshot, nil
	}

	// If not found, get global rating as baseline and create new snapshot
	var globalRating float64
	err = tx.QueryRowContext(ctx,
		`SELECT "Rating" FROM "Torrons" WHERE "Id" = $1`,
		torronId,
	).Scan(&globalRating)

	if err != nil {
		return nil, handleErrors(err)
	}

	// Create new snapshot with global rating as baseline
	newSnapshot := &domain.UserEloSnapshot{
		Id:        uuid.NewString(),
		UserId:    userId,
		TorronId:  torronId,
		Rating:    globalRating,
		VoteCount: 0,
		LastUpdated: time.Now().UTC().Format(time.RFC3339),
	}

	err = tx.QueryRowContext(ctx,
		`INSERT INTO "UserEloSnapshots" ("Id", "UserId", "TorronId", "Rating", "VoteCount", "LastUpdated")
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING "Id"`,
		newSnapshot.Id,
		newSnapshot.UserId,
		newSnapshot.TorronId,
		newSnapshot.Rating,
		newSnapshot.VoteCount,
		newSnapshot.LastUpdated,
	).Scan(&newSnapshot.Id)

	if err != nil {
		return nil, handleErrors(err)
	}

	return newSnapshot, nil
}

func (r *postgresUserEloSnapshotRepo) UpdateTx(tx *sql.Tx, ctx context.Context, snapshot *domain.UserEloSnapshot) (*domain.UserEloSnapshot, error) {
	snapshot.LastUpdated = time.Now().UTC().Format(time.RFC3339)

	_, err := tx.ExecContext(ctx,
		`UPDATE "UserEloSnapshots"
		 SET "Rating" = $3, "VoteCount" = $4, "LastUpdated" = $5
		 WHERE "UserId" = $1 AND "TorronId" = $2`,
		snapshot.UserId,
		snapshot.TorronId,
		snapshot.Rating,
		snapshot.VoteCount,
		snapshot.LastUpdated,
	)

	if err != nil {
		return nil, handleErrors(err)
	}

	return snapshot, nil
}
