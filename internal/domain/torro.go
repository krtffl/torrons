package domain

type Torro struct {
	Id     string  `db:"Id"     json:"id"`
	Name   string  `db:"Name"   json:"name"`
	Rating float64 `db:"Rating" json:"rating"`
	Image  string  `db:"Image"  json:"image"`
	Class  string  `db:"Class"  json:"class"`
}

type TorroRepo interface {
	List() ([]*Torro, error)
	ListByClass(classId string) ([]*Torro, error)
}
