package repository

import (
	"database/sql"
	"math/rand"

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

func (r *postgresPairingRepo) Get(id string) (*domain.Pairing, error) {
	row := r.db.QueryRow(
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

func (r *postgresPairingRepo) Create(pairing *domain.Pairing) (
	*domain.Pairing, error,
) {
	err := r.db.QueryRow(
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

func (r *postgresPairingRepo) List() ([]*domain.Pairing, error) {
	rows, err := r.db.Query(
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

func (r *postgresPairingRepo) ListByClass(classId string) ([]*domain.Pairing, error) {
	rows, err := r.db.Query(
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

func (r *postgresPairingRepo) GetRandom(classId string) (*domain.Pairing, error) {
	// Get count of pairings for this class
	count, err := r.CountClass(classId)
	if err != nil {
		return nil, err
	}

	// Handle edge case: no pairings available
	if count == 0 {
		return nil, handleErrors(sql.ErrNoRows)
	}

	// Generate random offset in application (more efficient than ORDER BY RANDOM())
	offset := rand.Intn(count)

	row := r.db.QueryRow(
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

func (r *postgresPairingRepo) Count() (int, error) {
	var count int
	err := r.db.QueryRow(
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

func (r *postgresPairingRepo) CountClass(classId string) (int, error) {
	var count int
	err := r.db.QueryRow(
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
