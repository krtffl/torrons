package domain

type Pairing struct {
	Id     string `db:"Id"`
	Torro1 string `db:"Torro1"`
	Torro2 string `db:"Torro2"`
	Class  string `db:"Class"`
}

type PairingRepo interface {
	Get(id string) (*Pairing, error)
	List() ([]*Pairing, error)
	ListByClass(classId string) ([]*Pairing, error)
	GetRandom(classId string) (*Pairing, error)
	Count() (int, error)
	CountClass(classId string) (int, error)
	Create(pairing *Pairing) (*Pairing, error)
}
