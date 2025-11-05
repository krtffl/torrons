-- Drop indexes first
DROP INDEX IF EXISTS idx_campaigns_dates;
DROP INDEX IF EXISTS idx_campaigns_year;
DROP INDEX IF EXISTS idx_campaigns_status;

-- Drop Campaigns table
DROP TABLE IF EXISTS "Campaigns";
