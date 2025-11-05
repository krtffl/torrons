package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresSessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) domain.SessionRepo {
	return &postgresSessionRepo{
		db: db,
	}
}

func (r *postgresSessionRepo) Create(session *domain.Session) (
	*domain.Session, error,
) {
	now := time.Now()
	err := r.db.QueryRow(
		`
        INSERT INTO "Sessions"
        ("Id", "CreatedAt", "LastSeenAt", "VoteCount", "Completed")
        VALUES
        ($1, $2, $3, $4, $5)
        RETURNING "Id", "CreatedAt", "LastSeenAt", "VoteCount", "Completed"`,
		uuid.NewString(),
		now,
		now,
		0,
		false,
	).Scan(&session.Id, &session.CreatedAt, &session.LastSeenAt, &session.VoteCount, &session.Completed)
	if err != nil {
		return nil, handleErrors(err)
	}

	return session, nil
}

func (r *postgresSessionRepo) Get(id string) (*domain.Session, error) {
	session := &domain.Session{}
	err := r.db.QueryRow(
		`SELECT "Id", "CreatedAt", "LastSeenAt", "VoteCount", "Completed"
         FROM "Sessions"
         WHERE "Id" = $1`,
		id,
	).Scan(&session.Id, &session.CreatedAt, &session.LastSeenAt, &session.VoteCount, &session.Completed)
	if err != nil {
		return nil, handleErrors(err)
	}

	return session, nil
}

func (r *postgresSessionRepo) Update(session *domain.Session) (
	*domain.Session, error,
) {
	err := r.db.QueryRow(
		`UPDATE "Sessions"
         SET "LastSeenAt" = $1, "VoteCount" = $2, "Completed" = $3
         WHERE "Id" = $4
         RETURNING "Id", "CreatedAt", "LastSeenAt", "VoteCount", "Completed"`,
		time.Now(),
		session.VoteCount,
		session.Completed,
		session.Id,
	).Scan(&session.Id, &session.CreatedAt, &session.LastSeenAt, &session.VoteCount, &session.Completed)
	if err != nil {
		return nil, handleErrors(err)
	}

	return session, nil
}

func (r *postgresSessionRepo) IncrementVoteCount(id string) error {
	_, err := r.db.Exec(
		`UPDATE "Sessions"
         SET "VoteCount" = "VoteCount" + 1, "LastSeenAt" = $1
         WHERE "Id" = $2`,
		time.Now(),
		id,
	)
	if err != nil {
		return handleErrors(err)
	}

	return nil
}

func (r *postgresSessionRepo) MarkCompleted(id string) error {
	_, err := r.db.Exec(
		`UPDATE "Sessions"
         SET "Completed" = true, "LastSeenAt" = $1
         WHERE "Id" = $2`,
		time.Now(),
		id,
	)
	if err != nil {
		return handleErrors(err)
	}

	return nil
}
