package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresAdventVoteRepo struct {
	db *sql.DB
}

func NewAdventVoteRepo(db *sql.DB) domain.AdventVoteRepo {
	return &postgresAdventVoteRepo{
		db: db,
	}
}

func (r *postgresAdventVoteRepo) HasVotedToday(ctx context.Context, userId string, voteDate string) (bool, error) {
	var exists bool

	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM "AdventVotes"
			WHERE "UserId" = $1 AND "VoteDate" = $2
		 )`,
		userId,
		voteDate,
	).Scan(&exists)

	if err != nil {
		return false, handleErrors(err)
	}

	return exists, nil
}

func (r *postgresAdventVoteRepo) CreateTx(tx *sql.Tx, ctx context.Context, vote *domain.AdventVote) (*domain.AdventVote, error) {
	if vote.Id == "" {
		vote.Id = uuid.NewString()
	}

	err := tx.QueryRowContext(ctx,
		`INSERT INTO "AdventVotes" ("Id", "UserId", "VoteDate", "PairingId")
		 VALUES ($1, $2, $3, $4)
		 RETURNING "Id"`,
		vote.Id,
		vote.UserId,
		vote.VoteDate,
		vote.PairingId,
	).Scan(&vote.Id)

	if err != nil {
		return nil, handleErrors(err)
	}

	return vote, nil
}
