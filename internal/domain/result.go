package domain

import (
	"context"
	"database/sql"
)

type Result struct {
	Id      string  `db:"Id"`
	Pairing string  `db:"Pairing"`
	Rat1Bef float64 `db:"Torro1RatingBefore"`
	Rat2Bef float64 `db:"Torro2RatingBefore"`
	Winner  string  `db:"Winner"`
	Rat1Aft float64 `db:"Torro1RatingAfter"`
	Rat2Aft float64 `db:"Torro2RatingAfter"`

	// User tracking and campaign association (added in migration 000010)
	UserId     *string `db:"UserId"     json:"user_id,omitempty"`
	Timestamp  string  `db:"Timestamp"  json:"timestamp"`
	CampaignId *string `db:"CampaignId" json:"campaign_id,omitempty"`
}

type ResultRepo interface {
	Create(ctx context.Context, result *Result) (*Result, error)
	// Transaction method
	CreateTx(tx *sql.Tx, ctx context.Context, result *Result) (*Result, error)
}
