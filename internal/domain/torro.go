package domain

import (
	"context"
	"database/sql"
)

type Torro struct {
	Id      string  `db:"Id"     json:"id"`
	Name    string  `db:"Name"   json:"name"`
	Rating  float64 `db:"Rating" json:"rating"`
	Image   string  `db:"Image"  json:"image"`
	Class   string  `db:"Class"  json:"class"`
	Pairing string  `db:"-"      json:"-"`
}

type TorroRepo interface {
	Get(ctx context.Context, id string) (*Torro, error)
	List(ctx context.Context) ([]*Torro, error)
	ListByClass(ctx context.Context, classId string) ([]*Torro, error)
	Update(ctx context.Context, id string, rating float64) (*Torro, error)
	// Transaction methods
	GetTx(tx *sql.Tx, ctx context.Context, id string) (*Torro, error)
	UpdateTx(tx *sql.Tx, ctx context.Context, id string, rating float64) (*Torro, error)
}
