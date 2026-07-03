-- Phase 2 - The knockout: single-elimination bracket per (Campaign, Class).
-- This is deliberately separate from the Phase 1 open-voting / ELO schema:
-- brackets do not touch "Torrons"."Rating", they only read it once at
-- seeding time. See internal/domain/bracket.go for the full mechanic.

-- Brackets: one knockout tournament for a given class within a campaign.
CREATE TABLE IF NOT EXISTS "Brackets" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_brackets PRIMARY KEY,
    "CampaignId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_campaign
        REFERENCES "Campaigns"("Id") ON DELETE CASCADE,
    "ClassId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_class
        REFERENCES "Classes"("Id") ON DELETE CASCADE,
    -- Target bracket size, always a power of two (e.g. 8, 16). The actual
    -- number of seeded entries may be lower if the class doesn't have
    -- enough torrons; the remaining top seeds receive byes.
    "Size" INT NOT NULL
        CONSTRAINT chk_bracket_size_positive CHECK ("Size" >= 2),
    "CurrentRound" INT NOT NULL DEFAULT 1,
    "Status" VARCHAR(20) NOT NULL DEFAULT 'in_progress'
        CHECK ("Status" IN ('in_progress', 'completed')),
    "ChampionId" VARCHAR(36)
        CONSTRAINT fk_bracket_champion
        REFERENCES "Torrons"("Id") ON DELETE SET NULL,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW(),
    "CompletedAt" TIMESTAMP
);

-- Only one active bracket per class per campaign.
CREATE UNIQUE INDEX idx_brackets_campaign_class ON "Brackets"("CampaignId", "ClassId");
CREATE INDEX idx_brackets_class ON "Brackets"("ClassId");
CREATE INDEX idx_brackets_status ON "Brackets"("Status");

-- BracketEntries: the seeded field for a bracket, seeded 1..N by Phase 1
-- ELO rating at the moment the bracket was created.
CREATE TABLE IF NOT EXISTS "BracketEntries" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_bracket_entries PRIMARY KEY,
    "BracketId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_entry_bracket
        REFERENCES "Brackets"("Id") ON DELETE CASCADE,
    "TorronId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_entry_torron
        REFERENCES "Torrons"("Id") ON DELETE CASCADE,
    "Seed" INT NOT NULL
        CONSTRAINT chk_bracket_entry_seed_positive CHECK ("Seed" >= 1),
    -- Rating at seeding time, kept for historical/share-card purposes.
    "SeedRating" NUMERIC NOT NULL
);

CREATE UNIQUE INDEX idx_bracket_entries_seed ON "BracketEntries"("BracketId", "Seed");
CREATE UNIQUE INDEX idx_bracket_entries_torron ON "BracketEntries"("BracketId", "TorronId");

-- BracketMatches: one match per (round, slot). Torro2 is NULL for a bye
-- (auto-advance, no vote needed). Winner is NULL until the match (or the
-- round it belongs to) has been decided.
CREATE TABLE IF NOT EXISTS "BracketMatches" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_bracket_matches PRIMARY KEY,
    "BracketId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_match_bracket
        REFERENCES "Brackets"("Id") ON DELETE CASCADE,
    "Round" INT NOT NULL
        CONSTRAINT chk_bracket_match_round_positive CHECK ("Round" >= 1),
    "Slot" INT NOT NULL
        CONSTRAINT chk_bracket_match_slot_non_negative CHECK ("Slot" >= 0),
    "Torro1Id" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_match_torro1
        REFERENCES "Torrons"("Id") ON DELETE CASCADE,
    "Torro2Id" VARCHAR(36)
        CONSTRAINT fk_bracket_match_torro2
        REFERENCES "Torrons"("Id") ON DELETE CASCADE,
    "WinnerId" VARCHAR(36)
        CONSTRAINT fk_bracket_match_winner
        REFERENCES "Torrons"("Id") ON DELETE SET NULL,
    "Status" VARCHAR(20) NOT NULL DEFAULT 'pending'
        CHECK ("Status" IN ('pending', 'completed')),
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_bracket_matches_slot ON "BracketMatches"("BracketId", "Round", "Slot");
CREATE INDEX idx_bracket_matches_round ON "BracketMatches"("BracketId", "Round");
CREATE INDEX idx_bracket_matches_status ON "BracketMatches"("Status");

-- BracketMatchVotes: per-user-per-match vote log. This is the DB-level
-- enforcement of the one-vote-per-match knockout rule (unlike Phase 1's
-- unlimited repeat voting on "Results").
CREATE TABLE IF NOT EXISTS "BracketMatchVotes" (
    "Id" VARCHAR(36) NOT NULL
        CONSTRAINT pk_bracket_match_votes PRIMARY KEY,
    "MatchId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_vote_match
        REFERENCES "BracketMatches"("Id") ON DELETE CASCADE,
    "UserId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_vote_user
        REFERENCES "Users"("Id") ON DELETE CASCADE,
    "TorronId" VARCHAR(36) NOT NULL
        CONSTRAINT fk_bracket_vote_torron
        REFERENCES "Torrons"("Id") ON DELETE CASCADE,
    "CreatedAt" TIMESTAMP NOT NULL DEFAULT NOW()
);

-- One vote per user per match - enforced at the DB level, not just in
-- application code.
CREATE UNIQUE INDEX idx_bracket_match_votes_unique ON "BracketMatchVotes"("MatchId", "UserId");
CREATE INDEX idx_bracket_match_votes_match ON "BracketMatchVotes"("MatchId");
