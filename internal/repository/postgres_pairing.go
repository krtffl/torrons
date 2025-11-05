package repository

import (
	"context"
	"crypto/rand"
	"database/sql"
	"math/big"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresPairingRepo struct {
	db *sql.DB
}

func NewPairingRepo(db *sql.DB) domain.PairingRepo {
	return &postgresPairingRepo{
		db: db,
	}
}

func (r *postgresPairingRepo) Get(ctx context.Context, id string) (*domain.Pairing, error) {
	row := r.db.QueryRowContext(ctx,
		`
        SELECT "Id", "Torro1", "Torro2", "Class"
        FROM "Pairings"
        WHERE "Id" = $1`,
		id,
	)
	pairing := &domain.Pairing{}
	err := row.Scan(
		&pairing.Id,
		&pairing.Torro1,
		&pairing.Torro2,
		&pairing.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return pairing, nil
}

func (r *postgresPairingRepo) Create(ctx context.Context, pairing *domain.Pairing) (
	*domain.Pairing, error,
) {
	err := r.db.QueryRowContext(ctx,
		`
        INSERT INTO "Pairings"
        ("Id", "Torro1", "Torro2", "Class")

        VALUES
        ($1, $2, $3, $4)
        RETURNING "Id"`,
		uuid.NewString(),
		pairing.Torro1,
		pairing.Torro2,
		pairing.Class,
	).Scan(&pairing.Id)
	if err != nil {
		return nil, handleErrors(err)
	}

	return pairing, nil
}

func (r *postgresPairingRepo) List(ctx context.Context) ([]*domain.Pairing, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Torro1", "Torro2", "Class"
        FROM "Pairings"`,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	defer rows.Close()
	var pairings []*domain.Pairing

	for rows.Next() {
		pairing := &domain.Pairing{}
		if err := rows.Scan(
			&pairing.Id,
			&pairing.Torro1,
			&pairing.Torro2,
			&pairing.Class,
		); err != nil {
			return nil, handleErrors(err)
		}
		pairings = append(pairings, pairing)
	}

	return pairings, nil
}

func (r *postgresPairingRepo) ListByClass(ctx context.Context, classId string) ([]*domain.Pairing, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Torro1", "Torro2", "Class"
        FROM "Pairings"
        WHERE "Class" = $1`,
		classId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	defer rows.Close()
	var pairings []*domain.Pairing

	for rows.Next() {
		pairing := &domain.Pairing{}
		if err := rows.Scan(
			&pairing.Id,
			&pairing.Torro1,
			&pairing.Torro2,
			&pairing.Class,
		); err != nil {
			return nil, handleErrors(err)
		}
		pairings = append(pairings, pairing)
	}

	return pairings, nil
}

func (r *postgresPairingRepo) GetRandom(ctx context.Context, classId string) (*domain.Pairing, error) {
	// Get count of pairings for this class
	count, err := r.CountClass(ctx, classId)
	if err != nil {
		return nil, err
	}

	// Handle edge case: no pairings available
	if count == 0 {
		return nil, handleErrors(sql.ErrNoRows)
	}

	// Generate cryptographically secure random offset
	// Using crypto/rand instead of math/rand for better security
	offsetBig, err := rand.Int(rand.Reader, big.NewInt(int64(count)))
	if err != nil {
		return nil, handleErrors(err)
	}
	offset := int(offsetBig.Int64())

	row := r.db.QueryRowContext(ctx,
		`
        SELECT "Id", "Torro1", "Torro2", "Class"
        FROM "Pairings"
        WHERE "Class" = $1
        LIMIT 1 OFFSET $2`,
		classId,
		offset,
	)
	pairing := &domain.Pairing{}
	err = row.Scan(
		&pairing.Id,
		&pairing.Torro1,
		&pairing.Torro2,
		&pairing.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return pairing, nil
}

func (r *postgresPairingRepo) Count(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`
        SELECT COUNT(*)
        FROM "Pairings"`,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, handleErrors(err)
	}

	return count, nil
}

func (r *postgresPairingRepo) CountClass(ctx context.Context, classId string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx,
		`
        SELECT COUNT(*)
        FROM "Pairings"
        WHERE "Class" = $1`,
		classId,
	).Scan(
		&count,
	)
	if err != nil {
		return 0, handleErrors(err)
	}

	return count, nil
}
