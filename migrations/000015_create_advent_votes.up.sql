-- Create AdventVotes table to gate the "advent daily duel" feature to one
-- vote per user per calendar day. The actual vote itself (ELO update, vote
-- count, streak, etc.) is recorded via the normal Results flow; this table
-- only tracks that a user has completed *today's* featured duel.
CREATE TABLE IF NOT EXISTS "AdventVotes" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_advent_votes PRIMARY KEY,
    "UserId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_advent_votes_user
        REFERENCES "Users"("Id") ON DELETE CASCADE,
    "VoteDate" DATE NOT NULL,
    "PairingId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_advent_votes_pairing
        REFERENCES "Pairings"("Id") ON DELETE CASCADE,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- One advent vote per user per day
CREATE UNIQUE INDEX idx_advent_votes_user_date ON "AdventVotes"("UserId", "VoteDate");
