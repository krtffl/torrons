package domain

import "database/sql"

type Torro struct {
	Id      string  `db:"Id"     json:"id"`
	Name    string  `db:"Name"   json:"name"`
	Rating  float64 `db:"Rating" json:"rating"`
	Image   string  `db:"Image"  json:"image"`
	Class   string  `db:"Class"  json:"class"`
	Pairing string  `db:"-"      json:"-"`
}

type TorroRepo interface {
	Get(id string) (*Torro, error)
	List() ([]*Torro, error)
	ListByClass(classId string) ([]*Torro, error)
	Update(id string, rating float64) (*Torro, error)
	// Transaction methods
	GetTx(tx *sql.Tx, id string) (*Torro, error)
	UpdateTx(tx *sql.Tx, id string, rating float64) (*Torro, error)
}
