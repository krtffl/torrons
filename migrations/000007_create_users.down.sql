-- Drop indexes first
DROP INDEX IF EXISTS idx_users_last_seen;
DROP INDEX IF EXISTS idx_users_vote_count;

-- Drop Users table
DROP TABLE IF EXISTS "Users";
