package domain

import (
	"context"
	"database/sql"
	"encoding/json"
)

// User represents an anonymous user tracked via cookie
type User struct {
	Id         string          `db:"Id"         json:"id"`
	FirstSeen  string          `db:"FirstSeen"  json:"first_seen"`
	LastSeen   string          `db:"LastSeen"   json:"last_seen"`
	VoteCount  int             `db:"VoteCount"  json:"vote_count"`
	ClassVotes json.RawMessage `db:"ClassVotes" json:"class_votes"` // JSONB: {"1": 15, "2": 30, ...}
}

// ClassVotesMap is a helper type for working with the ClassVotes JSONB field
type ClassVotesMap map[string]int

// UserRepo defines the interface for user data access
type UserRepo interface {
	// Get retrieves a user by ID
	Get(ctx context.Context, id string) (*User, error)

	// Create creates a new user
	Create(ctx context.Context, user *User) (*User, error)

	// Update updates user information
	Update(ctx context.Context, user *User) (*User, error)

	// IncrementVoteCount increments the user's total vote count and class-specific count
	IncrementVoteCount(ctx context.Context, userId string, classId string) error

	// GetVoteCountForClass gets the number of votes a user has cast in a specific class
	GetVoteCountForClass(ctx context.Context, userId string, classId string) (int, error)

	// UpdateLastSeen updates the user's last seen timestamp
	UpdateLastSeen(ctx context.Context, userId string) error

	// Transaction methods
	GetTx(tx *sql.Tx, ctx context.Context, id string) (*User, error)
	IncrementVoteCountTx(tx *sql.Tx, ctx context.Context, userId string, classId string) error
}
