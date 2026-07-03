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

	// Extended product information (added in migration 000011)
	Description     *string  `db:"Description"     json:"description,omitempty"`
	Weight          *string  `db:"Weight"          json:"weight,omitempty"`
	Price           *float64 `db:"Price"           json:"price,omitempty"`
	ProductUrl      *string  `db:"ProductUrl"      json:"product_url,omitempty"`
	Allergens       []string `db:"Allergens"       json:"allergens,omitempty"`
	MainIngredients []string `db:"MainIngredients" json:"main_ingredients,omitempty"`
	IsVegan         bool     `db:"IsVegan"         json:"is_vegan"`
	IsGlutenFree    bool     `db:"IsGlutenFree"    json:"is_gluten_free"`
	IsLactoseFree   bool     `db:"IsLactoseFree"   json:"is_lactose_free"`
	IsOrganic       bool     `db:"IsOrganic"       json:"is_organic"`
	IntensityLevel  *int     `db:"IntensityLevel"  json:"intensity_level,omitempty"`
	IsNew2025       bool     `db:"IsNew2025"       json:"is_new_2025"`
	Discontinued    bool     `db:"Discontinued"    json:"discontinued"`
	YearAdded       int      `db:"YearAdded"       json:"year_added"`
}

// TorroFilter holds optional dietary attribute filters used when listing
// torrons. The zero value (all fields false) means "no filtering applied".
type TorroFilter struct {
	IsVegan       bool
	IsGlutenFree  bool
	IsLactoseFree bool
	IsOrganic     bool
}

// IsEmpty reports whether no filter flags are set.
func (f TorroFilter) IsEmpty() bool {
	return !f.IsVegan && !f.IsGlutenFree && !f.IsLactoseFree && !f.IsOrganic
}

type TorroRepo interface {
	Get(ctx context.Context, id string) (*Torro, error)
	List(ctx context.Context) ([]*Torro, error)
	ListByClass(ctx context.Context, classId string) ([]*Torro, error)
	// ListFiltered lists torrons optionally scoped to a class and filtered by
	// dietary attributes. An empty classId returns torrons across all
	// classes. Results are ordered by rating (descending).
	ListFiltered(ctx context.Context, classId string, filter TorroFilter) ([]*Torro, error)
	Update(ctx context.Context, id string, rating float64) (*Torro, error)

	// TopNByClass returns the top N active (non-discontinued) torrons in a
	// class ordered by Rating descending. Used to seed a Phase 2 bracket
	// from Phase 1 ELO ratings. If fewer than n active torrons exist in
	// the class, all of them are returned.
	TopNByClass(ctx context.Context, classId string, n int) ([]*Torro, error)

	// Transaction methods
	GetTx(tx *sql.Tx, ctx context.Context, id string) (*Torro, error)
	UpdateTx(tx *sql.Tx, ctx context.Context, id string, rating float64) (*Torro, error)
}
