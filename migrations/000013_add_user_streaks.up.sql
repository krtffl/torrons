-- Add voting-streak tracking to Users table
-- Tracks consecutive days a user has cast at least one vote (any class),
-- independent of the per-class vote counts already tracked in ClassVotes.
ALTER TABLE "Users"
    ADD COLUMN IF NOT EXISTS "CurrentStreak" INT NOT NULL DEFAULT 0;

ALTER TABLE "Users"
    ADD COLUMN IF NOT EXISTS "LongestStreak" INT NOT NULL DEFAULT 0;

-- Date-only (no time component): the calendar day the user last voted on
ALTER TABLE "Users"
    ADD COLUMN IF NOT EXISTS "LastVoteDate" DATE;
