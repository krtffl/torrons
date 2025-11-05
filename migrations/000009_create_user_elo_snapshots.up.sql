-- Create UserEloSnapshots table for tracking personalized torron ratings
-- Each user maintains their own ELO view based on their voting history
CREATE TABLE IF NOT EXISTS "UserEloSnapshots" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_user_elo_snapshots PRIMARY KEY,
    "UserId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_user_elo_user
        REFERENCES "Users"("Id") ON DELETE CASCADE,
    "TorronId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_user_elo_torron
        REFERENCES "Torrons"("Id") ON DELETE CASCADE,
    "Rating" NUMERIC NOT NULL DEFAULT 1500,
    "VoteCount" INT NOT NULL DEFAULT 0,
    "LastUpdated" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Unique constraint: one snapshot per user per torron
CREATE UNIQUE INDEX idx_user_elo_unique ON "UserEloSnapshots"("UserId", "TorronId");

-- Index for faster user leaderboard queries
CREATE INDEX idx_user_elo_user_rating ON "UserEloSnapshots"("UserId", "Rating" DESC);

-- Index for faster torron lookup
CREATE INDEX idx_user_elo_torron ON "UserEloSnapshots"("TorronId");
