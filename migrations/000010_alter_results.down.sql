-- Drop indexes first
DROP INDEX IF EXISTS idx_results_user_timestamp;
DROP INDEX IF EXISTS idx_results_campaign;
DROP INDEX IF EXISTS idx_results_timestamp;
DROP INDEX IF EXISTS idx_results_user;

-- Drop columns
ALTER TABLE "Results" DROP COLUMN IF EXISTS "CampaignId";
ALTER TABLE "Results" DROP COLUMN IF EXISTS "Timestamp";
ALTER TABLE "Results" DROP COLUMN IF EXISTS "UserId";
