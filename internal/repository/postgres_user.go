package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/krtffl/torro/internal/domain"
)

type postgresUserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) domain.UserRepo {
	return &postgresUserRepo{
		db: db,
	}
}

func (r *postgresUserRepo) Get(ctx context.Context, id string) (*domain.User, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "FirstSeen", "LastSeen", "VoteCount", "ClassVotes"
		 FROM "Users"
		 WHERE "Id" = $1`,
		id,
	)

	user := &domain.User{}
	err := row.Scan(
		&user.Id,
		&user.FirstSeen,
		&user.LastSeen,
		&user.VoteCount,
		&user.ClassVotes,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return user, nil
}

func (r *postgresUserRepo) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Generate new UUID if not provided
	if user.Id == "" {
		user.Id = uuid.NewString()
	}

	// Set timestamps
	now := time.Now().UTC()
	if user.FirstSeen == "" {
		user.FirstSeen = now.Format(time.RFC3339)
	}
	if user.LastSeen == "" {
		user.LastSeen = now.Format(time.RFC3339)
	}

	// Initialize ClassVotes if empty
	if len(user.ClassVotes) == 0 {
		user.ClassVotes = json.RawMessage("{}")
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "Users" ("Id", "FirstSeen", "LastSeen", "VoteCount", "ClassVotes")
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING "Id"`,
		user.Id,
		user.FirstSeen,
		user.LastSeen,
		user.VoteCount,
		user.ClassVotes,
	).Scan(&user.Id)

	if err != nil {
		return nil, handleErrors(err)
	}

	return user, nil
}

func (r *postgresUserRepo) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Users"
		 SET "LastSeen" = $2, "VoteCount" = $3, "ClassVotes" = $4
		 WHERE "Id" = $1`,
		user.Id,
		user.LastSeen,
		user.VoteCount,
		user.ClassVotes,
	)

	if err != nil {
		return nil, handleErrors(err)
	}

	return user, nil
}

func (r *postgresUserRepo) IncrementVoteCount(ctx context.Context, userId string, classId string) error {
	// Use PostgreSQL JSONB operations to atomically increment class vote count
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Users"
		 SET "VoteCount" = "VoteCount" + 1,
		     "LastSeen" = $2,
		     "ClassVotes" = CASE
		         WHEN "ClassVotes"->$3 IS NULL THEN
		             jsonb_set("ClassVotes", ARRAY[$3], '1')
		         ELSE
		             jsonb_set("ClassVotes", ARRAY[$3],
		                 to_jsonb(CAST("ClassVotes"->>$3 AS INTEGER) + 1))
		     END
		 WHERE "Id" = $1`,
		userId,
		time.Now().UTC().Format(time.RFC3339),
		classId,
	)

	return handleErrors(err)
}

func (r *postgresUserRepo) GetVoteCountForClass(ctx context.Context, userId string, classId string) (int, error) {
	var count sql.NullInt64

	err := r.db.QueryRowContext(ctx,
		`SELECT CAST("ClassVotes"->>$2 AS INTEGER) as count
		 FROM "Users"
		 WHERE "Id" = $1`,
		userId,
		classId,
	).Scan(&count)

	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}
		return 0, handleErrors(err)
	}

	if !count.Valid {
		return 0, nil
	}

	return int(count.Int64), nil
}

func (r *postgresUserRepo) UpdateLastSeen(ctx context.Context, userId string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Users"
		 SET "LastSeen" = $2
		 WHERE "Id" = $1`,
		userId,
		time.Now().UTC().Format(time.RFC3339),
	)

	return handleErrors(err)
}

// Transaction methods

func (r *postgresUserRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.User, error) {
	row := tx.QueryRowContext(ctx,
		`SELECT "Id", "FirstSeen", "LastSeen", "VoteCount", "ClassVotes"
		 FROM "Users"
		 WHERE "Id" = $1`,
		id,
	)

	user := &domain.User{}
	err := row.Scan(
		&user.Id,
		&user.FirstSeen,
		&user.LastSeen,
		&user.VoteCount,
		&user.ClassVotes,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return user, nil
}

func (r *postgresUserRepo) IncrementVoteCountTx(tx *sql.Tx, ctx context.Context, userId string, classId string) error {
	// Use PostgreSQL JSONB operations to atomically increment class vote count
	_, err := tx.ExecContext(ctx,
		`UPDATE "Users"
		 SET "VoteCount" = "VoteCount" + 1,
		     "LastSeen" = $2,
		     "ClassVotes" = CASE
		         WHEN "ClassVotes"->$3 IS NULL THEN
		             jsonb_set("ClassVotes", ARRAY[$3], '1')
		         ELSE
		             jsonb_set("ClassVotes", ARRAY[$3],
		                 to_jsonb(CAST("ClassVotes"->>$3 AS INTEGER) + 1))
		     END
		 WHERE "Id" = $1`,
		userId,
		time.Now().UTC().Format(time.RFC3339),
		classId,
	)

	return handleErrors(err)
}

// handleArrays is a helper to convert PostgreSQL arrays to Go slices
func handleArrays(src interface{}) []string {
	if src == nil {
		return []string{}
	}

	arr, ok := src.(pq.StringArray)
	if !ok {
		return []string{}
	}

	return []string(arr)
}
