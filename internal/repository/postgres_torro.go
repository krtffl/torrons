package repository

import (
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

func (r *postgresTorroRepo) List() ([]*domain.Torro, error) {
	rows, err := r.db.Query(
		`
        SELECT * 
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

func (r *postgresTorroRepo) ListByClass(classId string) ([]*domain.Torro, error) {
	rows, err := r.db.Query(
		`
        SELECT * 
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

func (r *postgresTorroRepo) Get(id string) (*domain.Torro, error) {
	row := r.db.QueryRow(
		`
        SELECT * 
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

func (r *postgresTorroRepo) Update(id string, rating float64) (
	*domain.Torro, error,
) {
	updatedTorro := &domain.Torro{}

	err := r.db.QueryRow(
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

func (r *postgresTorroRepo) GetTx(tx *sql.Tx, id string) (*domain.Torro, error) {
	row := tx.QueryRow(
		`
        SELECT *
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

func (r *postgresTorroRepo) UpdateTx(tx *sql.Tx, id string, rating float64) (*domain.Torro, error) {
	updatedTorro := &domain.Torro{}

	err := tx.QueryRow(
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
