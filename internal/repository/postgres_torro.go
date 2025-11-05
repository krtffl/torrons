package repository

import (
	"context"
	"database/sql"

	"github.com/krtffl/torro/internal/domain"
)

type postgresTorroRepo struct {
	db *sql.DB
}

func NewTorroRepo(db *sql.DB) domain.TorroRepo {
	return &postgresTorroRepo{
		db: db,
	}
}

func (r *postgresTorroRepo) List(ctx context.Context) ([]*domain.Torro, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"`,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	defer rows.Close()
	var torrons []*domain.Torro

	for rows.Next() {
		torro := &domain.Torro{}
		if err := rows.Scan(
			&torro.Id,
			&torro.Name,
			&torro.Rating,
			&torro.Image,
			&torro.Class,
		); err != nil {
			return nil, handleErrors(err)
		}
		torrons = append(torrons, torro)
	}

	return torrons, nil
}

func (r *postgresTorroRepo) ListByClass(ctx context.Context, classId string) ([]*domain.Torro, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE "Class" = $1`,
		classId,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	defer rows.Close()
	var torrons []*domain.Torro

	for rows.Next() {
		torro := &domain.Torro{}
		if err := rows.Scan(
			&torro.Id,
			&torro.Name,
			&torro.Rating,
			&torro.Image,
			&torro.Class,
		); err != nil {
			return nil, handleErrors(err)
		}
		torrons = append(torrons, torro)
	}

	return torrons, nil
}

func (r *postgresTorroRepo) Get(ctx context.Context, id string) (*domain.Torro, error) {
	row := r.db.QueryRowContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE "Id" = $1`,
		id,
	)

	torro := &domain.Torro{}
	err := row.Scan(
		&torro.Id,
		&torro.Name,
		&torro.Rating,
		&torro.Image,
		&torro.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return torro, nil
}

func (r *postgresTorroRepo) Update(ctx context.Context, id string, rating float64) (
	*domain.Torro, error,
) {
	updatedTorro := &domain.Torro{}

	err := r.db.QueryRowContext(ctx,
		`
        UPDATE "Torrons" SET
        "Rating" = $2
        WHERE "Id" = $1
        RETURNING *`,
		id,
		rating,
	).Scan(
		&updatedTorro.Id,
		&updatedTorro.Name,
		&updatedTorro.Rating,
		&updatedTorro.Image,
		&updatedTorro.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return updatedTorro, nil
}

// Transaction methods

func (r *postgresTorroRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.Torro, error) {
	row := tx.QueryRowContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE "Id" = $1`,
		id,
	)

	torro := &domain.Torro{}
	err := row.Scan(
		&torro.Id,
		&torro.Name,
		&torro.Rating,
		&torro.Image,
		&torro.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return torro, nil
}

func (r *postgresTorroRepo) UpdateTx(tx *sql.Tx, ctx context.Context, id string, rating float64) (*domain.Torro, error) {
	updatedTorro := &domain.Torro{}

	err := tx.QueryRowContext(ctx,
		`
        UPDATE "Torrons" SET
        "Rating" = $2
        WHERE "Id" = $1
        RETURNING *`,
		id,
		rating,
	).Scan(
		&updatedTorro.Id,
		&updatedTorro.Name,
		&updatedTorro.Rating,
		&updatedTorro.Image,
		&updatedTorro.Class,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return updatedTorro, nil
}
