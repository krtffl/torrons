package domain

import (
	"context"
)

// Campaign represents a time-bound voting campaign
type Campaign struct {
	Id          string  `db:"Id"          json:"id"`
	Name        string  `db:"Name"        json:"name"`
	StartDate   string  `db:"StartDate"   json:"start_date"`
	EndDate     string  `db:"EndDate"     json:"end_date"`
	Status      string  `db:"Status"      json:"status"` // active, ended, archived
	Year        int     `db:"Year"        json:"year"`
	Description *string `db:"Description" json:"description,omitempty"`
	CreatedAt   string  `db:"CreatedAt"   json:"created_at"`
}

// CampaignStatus constants
const (
	CampaignStatusActive   = "active"
	CampaignStatusEnded    = "ended"
	CampaignStatusArchived = "archived"
)

// CampaignRepo defines the interface for campaign data access
type CampaignRepo interface {
	// Get retrieves a campaign by ID
	Get(ctx context.Context, id string) (*Campaign, error)

	// Create creates a new campaign
	Create(ctx context.Context, campaign *Campaign) (*Campaign, error)

	// Update updates campaign information
	Update(ctx context.Context, campaign *Campaign) (*Campaign, error)

	// List retrieves all campaigns
	List(ctx context.Context) ([]*Campaign, error)

	// GetActive retrieves the currently active campaign
	GetActive(ctx context.Context) (*Campaign, error)

	// GetByYear retrieves campaigns for a specific year
	GetByYear(ctx context.Context, year int) ([]*Campaign, error)

	// UpdateStatus updates the campaign status
	UpdateStatus(ctx context.Context, id string, status string) error
}
