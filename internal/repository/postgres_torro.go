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
