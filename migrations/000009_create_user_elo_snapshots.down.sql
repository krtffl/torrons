-- Drop indexes first
DROP INDEX IF EXISTS idx_user_elo_torron;
DROP INDEX IF EXISTS idx_user_elo_user_rating;
DROP INDEX IF EXISTS idx_user_elo_unique;

-- Drop UserEloSnapshots table
DROP TABLE IF EXISTS "UserEloSnapshots";
