package domain

import "time"

type UserVote struct {
	Id        string    `db:"Id"`
	SessionId string    `db:"SessionId"`
	PairingId string    `db:"PairingId"`
	WinnerId  string    `db:"WinnerId"`
	VotedAt   time.Time `db:"VotedAt"`
}

type UserVoteRepo interface {
	Create(*UserVote) (*UserVote, error)
	ListBySession(sessionId string) ([]*UserVote, error)
	CountBySession(sessionId string) (int, error)
}
