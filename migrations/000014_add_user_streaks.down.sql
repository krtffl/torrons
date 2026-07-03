-- Drop streak-tracking columns
ALTER TABLE "Users" DROP COLUMN IF EXISTS "LastVoteDate";
ALTER TABLE "Users" DROP COLUMN IF EXISTS "LongestStreak";
ALTER TABLE "Users" DROP COLUMN IF EXISTS "CurrentStreak";
