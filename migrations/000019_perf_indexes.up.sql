-- Add indexes that back the two hottest unindexed Results scans:
--   * /wrapped self-joins Results on "Pairing" (full scan per user).
--   * The press "VotesForTorro" lookup filters Results by "Winner".
-- Both columns are foreign keys but were never indexed (Postgres does not
-- auto-index FK columns), so these turn per-request sequential scans into
-- index lookups on the ever-growing Results table.

CREATE INDEX IF NOT EXISTS idx_results_pairing ON "Results"("Pairing");
CREATE INDEX IF NOT EXISTS idx_results_winner ON "Results"("Winner");
