-- Drop the performance indexes added in the up migration.
DROP INDEX IF EXISTS idx_results_pairing;
DROP INDEX IF EXISTS idx_results_winner;
