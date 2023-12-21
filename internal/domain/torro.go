package domain

type Torro struct {
	Id     string  `db:"Id"     json:"id"`
	Name   string  `db:"Name"   json:"name"`
	Rating float64 `db:"Rating" json:"rating"`
	Image  string  `db:"Image"  json:"image"`
}

type Pairing struct {
	Torro1 Torro
	Torro2 Torro
}
