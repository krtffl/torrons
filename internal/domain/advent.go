package domain

import (
	"context"
	"database/sql"
)

// AdventVote records that a user has completed the featured "advent daily
// duel" for a given calendar date. It exists purely to gate one vote per
// user per day; the vote itself (ELO update, vote count, streak, ...) is
// recorded through the normal Result/vote flow.
type AdventVote struct {
	Id        string `db:"Id"        json:"id"`
	UserId    string `db:"UserId"    json:"user_id"`
	VoteDate  string `db:"VoteDate"  json:"vote_date"` // date-only, e.g. "2025-12-06"
	PairingId string `db:"PairingId" json:"pairing_id"`
	CreatedAt string `db:"CreatedAt" json:"created_at"`
}

// AdventVoteRepo defines the interface for advent-duel gating data access
type AdventVoteRepo interface {
	// HasVotedToday reports whether the user has already completed the
	// advent duel for the given date (date-only, "2006-01-02" format)
	HasVotedToday(ctx context.Context, userId string, voteDate string) (bool, error)

	// CreateTx records today's advent vote as part of the same transaction
	// that records the underlying Result/ELO update
	CreateTx(tx *sql.Tx, ctx context.Context, vote *AdventVote) (*AdventVote, error)
}
