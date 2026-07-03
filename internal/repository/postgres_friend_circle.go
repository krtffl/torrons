package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresFriendCircleRepo struct {
	db *sql.DB
}

func NewFriendCircleRepo(db *sql.DB) domain.FriendCircleRepo {
	return &postgresFriendCircleRepo{
		db: db,
	}
}

// maxInviteCodeAttempts bounds the retry loop for the (astronomically
// unlikely) case where a freshly generated invite code collides with an
// existing one.
const maxInviteCodeAttempts = 5

func (r *postgresFriendCircleRepo) Create(ctx context.Context, ownerUserId string) (*domain.FriendCircle, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer tx.Rollback()

	circle := &domain.FriendCircle{
		Id:          uuid.NewString(),
		OwnerUserId: ownerUserId,
		CreatedAt:   time.Now().UTC().Format(time.RFC3339),
	}

	var insertErr error
	for attempt := 0; attempt < maxInviteCodeAttempts; attempt++ {
		code, genErr := generateInviteCode()
		if genErr != nil {
			return nil, handleErrors(genErr)
		}

		insertErr = tx.QueryRowContext(ctx,
			`INSERT INTO "FriendCircles" ("Id", "OwnerUserId", "InviteCode", "CreatedAt")
			 VALUES ($1, $2, $3, $4)
			 RETURNING "Id"`,
			circle.Id,
			circle.OwnerUserId,
			code,
			circle.CreatedAt,
		).Scan(&circle.Id)

		if insertErr == nil {
			circle.InviteCode = code
			break
		}

		if !strings.Contains(insertErr.Error(), "duplicate key") {
			return nil, handleErrors(insertErr)
		}
		// Invite code collision: loop and try again with a fresh code
	}

	if insertErr != nil {
		return nil, handleErrors(insertErr)
	}

	// The owner is automatically the first member of their own circle
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO "FriendCircleMembers" ("CircleId", "UserId")
		 VALUES ($1, $2)`,
		circle.Id,
		circle.OwnerUserId,
	); err != nil {
		return nil, handleErrors(err)
	}

	if err := tx.Commit(); err != nil {
		return nil, handleErrors(err)
	}

	return circle, nil
}

func (r *postgresFriendCircleRepo) Get(ctx context.Context, id string) (*domain.FriendCircle, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "OwnerUserId", "InviteCode", "CreatedAt"
		 FROM "FriendCircles"
		 WHERE "Id" = $1`,
		id,
	)

	circle := &domain.FriendCircle{}
	err := row.Scan(&circle.Id, &circle.OwnerUserId, &circle.InviteCode, &circle.CreatedAt)
	if err != nil {
		return nil, handleErrors(err)
	}

	return circle, nil
}

func (r *postgresFriendCircleRepo) GetByInviteCode(ctx context.Context, inviteCode string) (*domain.FriendCircle, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "OwnerUserId", "InviteCode", "CreatedAt"
		 FROM "FriendCircles"
		 WHERE "InviteCode" = $1`,
		inviteCode,
	)

	circle := &domain.FriendCircle{}
	err := row.Scan(&circle.Id, &circle.OwnerUserId, &circle.InviteCode, &circle.CreatedAt)
	if err != nil {
		return nil, handleErrors(err)
	}

	return circle, nil
}

func (r *postgresFriendCircleRepo) AddMember(ctx context.Context, circleId string, userId string) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO "FriendCircleMembers" ("CircleId", "UserId")
		 VALUES ($1, $2)
		 ON CONFLICT ("CircleId", "UserId") DO NOTHING`,
		circleId,
		userId,
	)

	return handleErrors(err)
}

func (r *postgresFriendCircleRepo) IsMember(ctx context.Context, circleId string, userId string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM "FriendCircleMembers"
			WHERE "CircleId" = $1 AND "UserId" = $2
		 )`,
		circleId,
		userId,
	).Scan(&exists)

	if err != nil {
		return false, handleErrors(err)
	}

	return exists, nil
}

func (r *postgresFriendCircleRepo) ListForUser(ctx context.Context, userId string) ([]*domain.FriendCircle, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT fc."Id", fc."OwnerUserId", fc."InviteCode", fc."CreatedAt"
		 FROM "FriendCircles" fc
		 INNER JOIN "FriendCircleMembers" fcm ON fcm."CircleId" = fc."Id"
		 WHERE fcm."UserId" = $1
		 ORDER BY fc."CreatedAt" DESC`,
		userId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var circles []*domain.FriendCircle
	for rows.Next() {
		circle := &domain.FriendCircle{}
		if err := rows.Scan(&circle.Id, &circle.OwnerUserId, &circle.InviteCode, &circle.CreatedAt); err != nil {
			return nil, handleErrors(err)
		}
		circles = append(circles, circle)
	}

	return circles, nil
}

// GetCircleLeaderboard mirrors postgres_user_elo_snapshot.go's
// GetUserLeaderboard join pattern (UserEloSnapshots joined to Torrons), but
// instead of filtering to a single user it joins to FriendCircleMembers and
// averages each torron's rating across every circle member who has a
// personalized snapshot for it (members who never rated a given torron
// simply don't contribute a data point for it).
func (r *postgresFriendCircleRepo) GetCircleLeaderboard(ctx context.Context, circleId string, classId string) ([]*domain.UserLeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			ues."TorronId",
			t."Name",
			t."Image",
			AVG(ues."Rating") as "AvgRating",
			SUM(ues."VoteCount") as "TotalVotes",
			RANK() OVER (ORDER BY AVG(ues."Rating") DESC) as "Rank"
		 FROM "UserEloSnapshots" ues
		 INNER JOIN "Torrons" t ON ues."TorronId" = t."Id"
		 INNER JOIN "FriendCircleMembers" fcm
		     ON fcm."UserId" = ues."UserId" AND fcm."CircleId" = $1
		 WHERE t."Class" = $2
		   AND t."Discontinued" = false
		 GROUP BY ues."TorronId", t."Name", t."Image"
		 ORDER BY "AvgRating" DESC`,
		circleId,
		classId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanCircleLeaderboardEntries(rows)
}

// GetCircleGlobalLeaderboard is the circle-scoped equivalent of
// GetUserGlobalLeaderboard: same averaging as GetCircleLeaderboard, but
// across every class, capped to the top 100.
func (r *postgresFriendCircleRepo) GetCircleGlobalLeaderboard(ctx context.Context, circleId string) ([]*domain.UserLeaderboardEntry, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT
			ues."TorronId",
			t."Name",
			t."Image",
			AVG(ues."Rating") as "AvgRating",
			SUM(ues."VoteCount") as "TotalVotes",
			RANK() OVER (ORDER BY AVG(ues."Rating") DESC) as "Rank"
		 FROM "UserEloSnapshots" ues
		 INNER JOIN "Torrons" t ON ues."TorronId" = t."Id"
		 INNER JOIN "FriendCircleMembers" fcm
		     ON fcm."UserId" = ues."UserId" AND fcm."CircleId" = $1
		 WHERE t."Discontinued" = false
		 GROUP BY ues."TorronId", t."Name", t."Image"
		 ORDER BY "AvgRating" DESC
		 LIMIT 100`,
		circleId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	return scanCircleLeaderboardEntries(rows)
}

// scanCircleLeaderboardEntries scans rows shaped
// (TorronId, Name, Image, AvgRating, TotalVotes, Rank)
func scanCircleLeaderboardEntries(rows *sql.Rows) ([]*domain.UserLeaderboardEntry, error) {
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

// generateInviteCode returns a short, URL-safe, cryptographically random
// invite code (10 hex characters from 5 random bytes).
func generateInviteCode() (string, error) {
	buf := make([]byte, 5)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
