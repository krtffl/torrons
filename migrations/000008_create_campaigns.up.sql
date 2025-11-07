-- Create Campaigns table for managing time-bound voting campaigns
-- Each campaign has a start/end date and countdown to results reveal
CREATE TABLE IF NOT EXISTS "Campaigns" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_campaigns PRIMARY KEY,
    "Name" VARCHAR(255) NOT NULL,
    "StartDate" TIMESTAMP NOT NULL,
    "EndDate" TIMESTAMP NOT NULL,
    "Status" VARCHAR(20) NOT NULL DEFAULT 'active'
        CHECK ("Status" IN ('active', 'ended', 'archived')),
    "Year" INT NOT NULL,
    "Description" TEXT,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Index for faster lookups by status
CREATE INDEX idx_campaigns_status ON "Campaigns"("Status");

-- Index for faster lookups by year
CREATE INDEX idx_campaigns_year ON "Campaigns"("Year");

-- Index for finding active campaigns by date
CREATE INDEX idx_campaigns_dates ON "Campaigns"("StartDate", "EndDate");
