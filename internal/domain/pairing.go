package domain

type Pairing struct {
	Id     string
	Torro1 string
	Torro2 string
	Class  string
}

type PairingRepo interface {
	List() ([]*Pairing, error)
	ListByClass(classId string) ([]*Pairing, error)
	Count() (int, error)
	CountClass(classId string) (int, error)
	Create(pairing *Pairing) (*Pairing, error)
}
