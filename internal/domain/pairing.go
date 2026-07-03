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

	// GetDeterministic returns the same pairing for a given (classId, seed)
	// pair every time it's called, mirroring GetRandom's offset-based query
	// but with a caller-supplied seed instead of crypto/rand. Used to pick a
	// stable "pairing of the day" across requests and replicas.
	GetDeterministic(ctx context.Context, classId string, seed int64) (*Pairing, error)

	Count(ctx context.Context) (int, error)
	CountClass(ctx context.Context, classId string) (int, error)
	Create(ctx context.Context, pairing *Pairing) (*Pairing, error)
}
