package domain

import (
	"context"
	"database/sql"
)

// UserEloSnapshot represents a user's personalized ELO rating for a specific torron
// Each user maintains their own view of torron ratings based on their voting history
type UserEloSnapshot struct {
	Id          string  `db:"Id"          json:"id"`
	UserId      string  `db:"UserId"      json:"user_id"`
	TorronId    string  `db:"TorronId"    json:"torron_id"`
	Rating      float64 `db:"Rating"      json:"rating"`
	VoteCount   int     `db:"VoteCount"   json:"vote_count"`
	LastUpdated string  `db:"LastUpdated" json:"last_updated"`
}

// UserLeaderboardEntry represents a torron with its user-specific rating
// Used for generating personalized leaderboards
type UserLeaderboardEntry struct {
	TorronId    string  `json:"torron_id"`
	TorronName  string  `json:"torron_name"`
	TorronImage string  `json:"torron_image"`
	Rating      float64 `json:"rating"`
	VoteCount   int     `json:"vote_count"`
	Rank        int     `json:"rank"`
}

// UserEloSnapshotRepo defines the interface for user ELO snapshot data access
type UserEloSnapshotRepo interface {
	// Get retrieves a user's ELO snapshot for a specific torron
	Get(ctx context.Context, userId string, torronId string) (*UserEloSnapshot, error)

	// Create creates a new user ELO snapshot (initial rating)
	Create(ctx context.Context, snapshot *UserEloSnapshot) (*UserEloSnapshot, error)

	// Update updates a user's ELO snapshot after a vote
	Update(ctx context.Context, snapshot *UserEloSnapshot) (*UserEloSnapshot, error)

	// GetOrCreate retrieves existing snapshot or creates new one with default rating
	GetOrCreate(ctx context.Context, userId string, torronId string) (*UserEloSnapshot, error)

	// GetUserLeaderboard retrieves a user's personalized leaderboard for a class
	GetUserLeaderboard(ctx context.Context, userId string, classId string) ([]*UserLeaderboardEntry, error)

	// GetUserGlobalLeaderboard retrieves a user's personalized global leaderboard
	GetUserGlobalLeaderboard(ctx context.Context, userId string) ([]*UserLeaderboardEntry, error)

	// ListByUser retrieves all ELO snapshots for a user
	ListByUser(ctx context.Context, userId string) ([]*UserEloSnapshot, error)

	// Transaction methods
	GetTx(tx *sql.Tx, ctx context.Context, userId string, torronId string) (*UserEloSnapshot, error)
	GetOrCreateTx(tx *sql.Tx, ctx context.Context, userId string, torronId string) (*UserEloSnapshot, error)
	UpdateTx(tx *sql.Tx, ctx context.Context, snapshot *UserEloSnapshot) (*UserEloSnapshot, error)
}
