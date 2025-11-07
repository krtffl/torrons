package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/google/uuid"

	"github.com/krtffl/torro/internal/domain"
)

type postgresCampaignRepo struct {
	db *sql.DB
}

func NewCampaignRepo(db *sql.DB) domain.CampaignRepo {
	return &postgresCampaignRepo{
		db: db,
	}
}

func (r *postgresCampaignRepo) Get(ctx context.Context, id string) (*domain.Campaign, error) {
	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "Name", "StartDate", "EndDate", "Status", "Year", "Description", "CreatedAt"
		 FROM "Campaigns"
		 WHERE "Id" = $1`,
		id,
	)

	campaign := &domain.Campaign{}
	err := row.Scan(
		&campaign.Id,
		&campaign.Name,
		&campaign.StartDate,
		&campaign.EndDate,
		&campaign.Status,
		&campaign.Year,
		&campaign.Description,
		&campaign.CreatedAt,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return campaign, nil
}

func (r *postgresCampaignRepo) Create(ctx context.Context, campaign *domain.Campaign) (*domain.Campaign, error) {
	if campaign.Id == "" {
		campaign.Id = uuid.NewString()
	}

	if campaign.Status == "" {
		campaign.Status = domain.CampaignStatusActive
	}

	if campaign.CreatedAt == "" {
		campaign.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	err := r.db.QueryRowContext(ctx,
		`INSERT INTO "Campaigns" ("Id", "Name", "StartDate", "EndDate", "Status", "Year", "Description", "CreatedAt")
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING "Id"`,
		campaign.Id,
		campaign.Name,
		campaign.StartDate,
		campaign.EndDate,
		campaign.Status,
		campaign.Year,
		campaign.Description,
		campaign.CreatedAt,
	).Scan(&campaign.Id)

	if err != nil {
		return nil, handleErrors(err)
	}

	return campaign, nil
}

func (r *postgresCampaignRepo) Update(ctx context.Context, campaign *domain.Campaign) (*domain.Campaign, error) {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Campaigns"
		 SET "Name" = $2, "StartDate" = $3, "EndDate" = $4, "Status" = $5, "Year" = $6, "Description" = $7
		 WHERE "Id" = $1`,
		campaign.Id,
		campaign.Name,
		campaign.StartDate,
		campaign.EndDate,
		campaign.Status,
		campaign.Year,
		campaign.Description,
	)

	if err != nil {
		return nil, handleErrors(err)
	}

	return campaign, nil
}

func (r *postgresCampaignRepo) List(ctx context.Context) ([]*domain.Campaign, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "Name", "StartDate", "EndDate", "Status", "Year", "Description", "CreatedAt"
		 FROM "Campaigns"
		 ORDER BY "StartDate" DESC`,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var campaigns []*domain.Campaign
	for rows.Next() {
		campaign := &domain.Campaign{}
		err := rows.Scan(
			&campaign.Id,
			&campaign.Name,
			&campaign.StartDate,
			&campaign.EndDate,
			&campaign.Status,
			&campaign.Year,
			&campaign.Description,
			&campaign.CreatedAt,
		)
		if err != nil {
			return nil, handleErrors(err)
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

func (r *postgresCampaignRepo) GetActive(ctx context.Context) (*domain.Campaign, error) {
	now := time.Now().UTC().Format(time.RFC3339)

	row := r.db.QueryRowContext(ctx,
		`SELECT "Id", "Name", "StartDate", "EndDate", "Status", "Year", "Description", "CreatedAt"
		 FROM "Campaigns"
		 WHERE "Status" = $1
		   AND "StartDate" <= $2
		   AND "EndDate" >= $2
		 ORDER BY "StartDate" DESC
		 LIMIT 1`,
		domain.CampaignStatusActive,
		now,
	)

	campaign := &domain.Campaign{}
	err := row.Scan(
		&campaign.Id,
		&campaign.Name,
		&campaign.StartDate,
		&campaign.EndDate,
		&campaign.Status,
		&campaign.Year,
		&campaign.Description,
		&campaign.CreatedAt,
	)
	if err != nil {
		return nil, handleErrors(err)
	}

	return campaign, nil
}

func (r *postgresCampaignRepo) GetByYear(ctx context.Context, year int) ([]*domain.Campaign, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT "Id", "Name", "StartDate", "EndDate", "Status", "Year", "Description", "CreatedAt"
		 FROM "Campaigns"
		 WHERE "Year" = $1
		 ORDER BY "StartDate" DESC`,
		year,
	)
	if err != nil {
		return nil, handleErrors(err)
	}
	defer rows.Close()

	var campaigns []*domain.Campaign
	for rows.Next() {
		campaign := &domain.Campaign{}
		err := rows.Scan(
			&campaign.Id,
			&campaign.Name,
			&campaign.StartDate,
			&campaign.EndDate,
			&campaign.Status,
			&campaign.Year,
			&campaign.Description,
			&campaign.CreatedAt,
		)
		if err != nil {
			return nil, handleErrors(err)
		}
		campaigns = append(campaigns, campaign)
	}

	return campaigns, nil
}

func (r *postgresCampaignRepo) UpdateStatus(ctx context.Context, id string, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE "Campaigns"
		 SET "Status" = $2
		 WHERE "Id" = $1`,
		id,
		status,
	)

	return handleErrors(err)
}
