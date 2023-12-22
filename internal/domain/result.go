package domain

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
}
