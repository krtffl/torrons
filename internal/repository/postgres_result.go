package repository

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresResultRepo struct {
	db *sql.DB
}

func NewResultRepo(db *sql.DB) domain.ResultRepo {
	return &postgresResultRepo{
		db: db,
	}
}

func (r *postgresResultRepo) Create(ctx context.Context, result *domain.Result) (
	*domain.Result, error,
) {
	err := r.db.QueryRowContext(ctx,
		`
        INSERT INTO "Results"
        ("Id", "Pairing", "Torro1RatingBefore", "Torro2RatingBefore",
        "Winner", "Torro1RatingAfter", "Torro2RatingAfter")
        VALUES
        ($1, $2, $3, $4, $5, $6, $7)
        RETURNING "Id"`,
		uuid.NewString(),
		result.Pairing,
		result.Rat1Bef,
		result.Rat2Bef,
		result.Winner,
		result.Rat1Aft,
		result.Rat2Aft,
	).Scan(&result.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return result, nil
}

// Transaction method

func (r *postgresResultRepo) CreateTx(tx *sql.Tx, ctx context.Context, result *domain.Result) (
	*domain.Result, error,
) {
	err := tx.QueryRowContext(ctx,
		`
        INSERT INTO "Results"
        ("Id", "Pairing", "Torro1RatingBefore", "Torro2RatingBefore",
        "Winner", "Torro1RatingAfter", "Torro2RatingAfter")
        VALUES
        ($1, $2, $3, $4, $5, $6, $7)
        RETURNING "Id"`,
		uuid.NewString(),
		result.Pairing,
		result.Rat1Bef,
		result.Rat2Bef,
		result.Winner,
		result.Rat1Aft,
		result.Rat2Aft,
	).Scan(&result.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return result, nil
}
