package domain

type Class struct {
	Id          string `db:"Id"          json:"id"`
	Name        string `db:"Name"        json:"name"`
	Description string `db:"Description" json:"description"`
}

type ClassRepo interface {
	List() ([]*Class, error)
}
