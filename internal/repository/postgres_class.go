package repository

import (
	"context"
	"database/sql"

	"github.com/krtffl/torro/internal/domain"
)

type postgresClassRepo struct {
	db *sql.DB
}

func NewClassRepo(db *sql.DB) domain.ClassRepo {
	return &postgresClassRepo{
		db: db,
	}
}

func (r *postgresClassRepo) List(ctx context.Context) ([]*domain.Class, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Name", "Description"
        FROM "Classes"
        ORDER BY "Id" ASC`,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	defer rows.Close()
	var classes []*domain.Class

	for rows.Next() {
		class := &domain.Class{}
		if err := rows.Scan(
			&class.Id,
			&class.Name,
			&class.Description,
		); err != nil {
			return nil, handleErrors(err)
		}
		classes = append(classes, class)
	}

	return classes, nil
}
