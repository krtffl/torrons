package repository

import (
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresUserVoteRepo struct {
	db *sql.DB
}

func NewUserVoteRepo(db *sql.DB) domain.UserVoteRepo {
	return &postgresUserVoteRepo{
		db: db,
	}
}

func (r *postgresUserVoteRepo) Create(vote *domain.UserVote) (
	*domain.UserVote, error,
) {
	err := r.db.QueryRow(
		`
        INSERT INTO "UserVotes"
        ("Id", "SessionId", "PairingId", "WinnerId", "VotedAt")
        VALUES
        ($1, $2, $3, $4, $5)
        RETURNING "Id", "VotedAt"`,
		uuid.NewString(),
		vote.SessionId,
		vote.PairingId,
		vote.WinnerId,
		time.Now(),
	).Scan(&vote.Id, &vote.VotedAt)
	if err != nil {
		return nil, handleErrors(err)
	}

	return vote, nil
}

func (r *postgresUserVoteRepo) ListBySession(sessionId string) ([]*domain.UserVote, error) {
	rows, err := r.db.Query(
		`SELECT "Id", "SessionId", "PairingId", "WinnerId", "VotedAt"
         FROM "UserVotes"
         WHERE "SessionId" = $1
         ORDER BY "VotedAt" DESC`,
		sessionId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	votes := []*domain.UserVote{}
	for rows.Next() {
		vote := &domain.UserVote{}
		err := rows.Scan(&vote.Id, &vote.SessionId, &vote.PairingId, &vote.WinnerId, &vote.VotedAt)
		if err != nil {
			return nil, handleErrors(err)
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

func (r *postgresUserVoteRepo) CountBySession(sessionId string) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM "UserVotes" WHERE "SessionId" = $1`,
		sessionId,
	).Scan(&count)
	if err != nil {
		return 0, handleErrors(err)
	}

	return count, nil
}
