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

type TorroRepo interface {
	Get(ctx context.Context, id string) (*Torro, error)
	List(ctx context.Context) ([]*Torro, error)
	ListByClass(ctx context.Context, classId string) ([]*Torro, error)
	Update(ctx context.Context, id string, rating float64) (*Torro, error)
	// Transaction methods
	GetTx(tx *sql.Tx, ctx context.Context, id string) (*Torro, error)
	UpdateTx(tx *sql.Tx, ctx context.Context, id string, rating float64) (*Torro, error)
}
