package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/krtffl/torro/internal/domain"
	"github.com/lib/pq"
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
        SELECT "Id", "Name", "Rating", "Image", "Class",
               "Description", "Weight", "Price", "ProductUrl",
               "Allergens", "MainIngredients",
               "IsVegan", "IsGlutenFree", "IsLactoseFree", "IsOrganic",
               "IntensityLevel", "IsNew2025", "Discontinued", "YearAdded"
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
		&torro.Description,
		&torro.Weight,
		&torro.Price,
		&torro.ProductUrl,
		pq.Array(&torro.Allergens),
		pq.Array(&torro.MainIngredients),
		&torro.IsVegan,
		&torro.IsGlutenFree,
		&torro.IsLactoseFree,
		&torro.IsOrganic,
		&torro.IntensityLevel,
		&torro.IsNew2025,
		&torro.Discontinued,
		&torro.YearAdded,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return torro, nil
}

// ListFiltered lists torrons optionally scoped to a class and filtered by
// dietary attributes. An empty classId returns torrons across all classes.
// Results are ordered by rating (descending) so callers can rely on
// positional ranking.
func (r *postgresTorroRepo) ListFiltered(ctx context.Context, classId string, filter domain.TorroFilter) ([]*domain.Torro, error) {
	query := `
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE 1 = 1`
	var args []interface{}

	if classId != "" {
		args = append(args, classId)
		query += fmt.Sprintf(` AND "Class" = $%d`, len(args))
	}

	query += dietaryFilterSQL(filter, "")
	query += `
        ORDER BY "Rating" DESC`

	rows, err := r.db.QueryContext(ctx, query, args...)
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

// dietaryFilterSQL builds SQL "AND" conditions for the given dietary filter
// flags. alias, if non-empty, is used to qualify the column names (e.g.
// "t." when the "Torrons" table is joined under an alias).
func dietaryFilterSQL(filter domain.TorroFilter, alias string) string {
	var b strings.Builder
	if filter.IsVegan {
		fmt.Fprintf(&b, ` AND %s"IsVegan" = true`, alias)
	}
	if filter.IsGlutenFree {
		fmt.Fprintf(&b, ` AND %s"IsGlutenFree" = true`, alias)
	}
	if filter.IsLactoseFree {
		fmt.Fprintf(&b, ` AND %s"IsLactoseFree" = true`, alias)
	}
	if filter.IsOrganic {
		fmt.Fprintf(&b, ` AND %s"IsOrganic" = true`, alias)
	}
	return b.String()
}

// TopNByClass returns the top N active torrons in a class ordered by
// Rating descending, for seeding a Phase 2 bracket.
func (r *postgresTorroRepo) TopNByClass(ctx context.Context, classId string, n int) ([]*domain.Torro, error) {
	rows, err := r.db.QueryContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE "Class" = $1
          AND "Discontinued" = false
        ORDER BY "Rating" DESC
        LIMIT $2`,
		classId,
		n,
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

func (r *postgresTorroRepo) Update(ctx context.Context, id string, rating float64) (
	*domain.Torro, error,
) {
	updatedTorro := &domain.Torro{}

	err := r.db.QueryRowContext(ctx,
		`
        UPDATE "Torrons" SET
        "Rating" = $2
        WHERE "Id" = $1
        RETURNING "Id", "Name", "Rating", "Image", "Class",
                  "Description", "Weight", "Price", "ProductUrl",
                  "Allergens", "MainIngredients",
                  "IsVegan", "IsGlutenFree", "IsLactoseFree", "IsOrganic",
                  "IntensityLevel", "IsNew2025", "Discontinued", "YearAdded"`,
		id,
		rating,
	).Scan(
		&updatedTorro.Id,
		&updatedTorro.Name,
		&updatedTorro.Rating,
		&updatedTorro.Image,
		&updatedTorro.Class,
		&updatedTorro.Description,
		&updatedTorro.Weight,
		&updatedTorro.Price,
		&updatedTorro.ProductUrl,
		pq.Array(&updatedTorro.Allergens),
		pq.Array(&updatedTorro.MainIngredients),
		&updatedTorro.IsVegan,
		&updatedTorro.IsGlutenFree,
		&updatedTorro.IsLactoseFree,
		&updatedTorro.IsOrganic,
		&updatedTorro.IntensityLevel,
		&updatedTorro.IsNew2025,
		&updatedTorro.Discontinued,
		&updatedTorro.YearAdded,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return updatedTorro, nil
}

// Transaction methods

// GetTx reads a torró row inside a transaction and takes a FOR UPDATE row
// lock. It is used exclusively by the vote path (Handler.result), which does a
// read-modify-write of "Rating"; the lock serializes concurrent votes on the
// same torró so they can no longer read the same pre-image and clobber each
// other's rating update (lost update). Callers must acquire these locks in a
// deterministic order (sorted by id) to avoid deadlocks.
func (r *postgresTorroRepo) GetTx(tx *sql.Tx, ctx context.Context, id string) (*domain.Torro, error) {
	row := tx.QueryRowContext(ctx,
		`
        SELECT "Id", "Name", "Rating", "Image", "Class"
        FROM "Torrons"
        WHERE "Id" = $1
        FOR UPDATE`,
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
        RETURNING "Id", "Name", "Rating", "Image", "Class",
                  "Description", "Weight", "Price", "ProductUrl",
                  "Allergens", "MainIngredients",
                  "IsVegan", "IsGlutenFree", "IsLactoseFree", "IsOrganic",
                  "IntensityLevel", "IsNew2025", "Discontinued", "YearAdded"`,
		id,
		rating,
	).Scan(
		&updatedTorro.Id,
		&updatedTorro.Name,
		&updatedTorro.Rating,
		&updatedTorro.Image,
		&updatedTorro.Class,
		&updatedTorro.Description,
		&updatedTorro.Weight,
		&updatedTorro.Price,
		&updatedTorro.ProductUrl,
		pq.Array(&updatedTorro.Allergens),
		pq.Array(&updatedTorro.MainIngredients),
		&updatedTorro.IsVegan,
		&updatedTorro.IsGlutenFree,
		&updatedTorro.IsLactoseFree,
		&updatedTorro.IsOrganic,
		&updatedTorro.IntensityLevel,
		&updatedTorro.IsNew2025,
		&updatedTorro.Discontinued,
		&updatedTorro.YearAdded,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	return updatedTorro, nil
}
