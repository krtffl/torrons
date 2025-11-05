package domain

import "time"

type Session struct {
	Id         string    `db:"Id"`
	CreatedAt  time.Time `db:"CreatedAt"`
	LastSeenAt time.Time `db:"LastSeenAt"`
	VoteCount  int       `db:"VoteCount"`
	Completed  bool      `db:"Completed"`
}

type SessionRepo interface {
	Create(*Session) (*Session, error)
	Get(id string) (*Session, error)
	Update(*Session) (*Session, error)
	IncrementVoteCount(id string) error
	MarkCompleted(id string) error
}
