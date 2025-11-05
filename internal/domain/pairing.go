package domain

import "context"

type Pairing struct {
	Id     string `db:"Id"`
	Torro1 string `db:"Torro1"`
	Torro2 string `db:"Torro2"`
	Class  string `db:"Class"`
}

type PairingRepo interface {
	Get(ctx context.Context, id string) (*Pairing, error)
	List(ctx context.Context) ([]*Pairing, error)
	ListByClass(ctx context.Context, classId string) ([]*Pairing, error)
	GetRandom(ctx context.Context, classId string) (*Pairing, error)
	Count(ctx context.Context) (int, error)
	CountClass(ctx context.Context, classId string) (int, error)
	Create(ctx context.Context, pairing *Pairing) (*Pairing, error)
}
