-- Create Users table for tracking anonymous users via cookies
-- Users are identified by UUID stored in browser cookie
CREATE TABLE IF NOT EXISTS "Users" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_users PRIMARY KEY,
    "FirstSeen" TIMESTAMP NOT NULL DEFAULT NOW(),
    "LastSeen" TIMESTAMP NOT NULL DEFAULT NOW(),
    "VoteCount" INT NOT NULL DEFAULT 0,
    "ClassVotes" JSONB DEFAULT '{}'::jsonb
);

-- Index for faster lookups by vote count
CREATE INDEX idx_users_vote_count ON "Users"("VoteCount");

-- Index for faster lookups by last seen
CREATE INDEX idx_users_last_seen ON "Users"("LastSeen");
