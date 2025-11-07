-- Add user tracking and timestamp to Results table
-- This enables vote history analysis and campaign association

-- Add UserId column (nullable for backward compatibility with existing data)
ALTER TABLE "Results"
    ADD COLUMN IF NOT EXISTS "UserId" VARCHAR(36)
        CONSTRAINT fk_result_user
        REFERENCES "Users"("Id") ON DELETE SET NULL;

-- Add Timestamp column (with default for existing rows)
ALTER TABLE "Results"
    ADD COLUMN IF NOT EXISTS "Timestamp" TIMESTAMP NOT NULL DEFAULT NOW();

-- Add CampaignId column (nullable for votes outside campaigns)
ALTER TABLE "Results"
    ADD COLUMN IF NOT EXISTS "CampaignId" VARCHAR(36)
        CONSTRAINT fk_result_campaign
        REFERENCES "Campaigns"("Id") ON DELETE SET NULL;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_results_user ON "Results"("UserId");
CREATE INDEX IF NOT EXISTS idx_results_timestamp ON "Results"("Timestamp" DESC);
CREATE INDEX IF NOT EXISTS idx_results_campaign ON "Results"("CampaignId");
CREATE INDEX IF NOT EXISTS idx_results_user_timestamp ON "Results"("UserId", "Timestamp" DESC);
