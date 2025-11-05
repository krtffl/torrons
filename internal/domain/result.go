package domain

import "database/sql"

type Result struct {
	Id      string  `db:"Id"`
	Pairing string  `db:"Pairing"`
	Rat1Bef float64 `db:"Torro1RatingBefore"`
	Rat2Bef float64 `db:"Torro2RatingBefore"`
	Winner  string  `db:"Winner"`
	Rat1Aft float64 `db:"Torro1RatingAfter"`
	Rat2Aft float64 `db:"Torro2RatingAfter"`
}

type ResultRepo interface {
	Create(*Result) (*Result, error)
	// Transaction method
	CreateTx(tx *sql.Tx, result *Result) (*Result, error)
}
