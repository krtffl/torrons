-- Prevent duplicate pairings: a multi-instance boot race where two
-- processes each see zero existing pairings for a class and both seed
-- their own full set concurrently (internal/api.CheckPairingsCreated) can
-- otherwise double-insert the same matchup. Order-independent - a pairing
-- stored as (A,B) and one stored as (B,A) for the same class are the same
-- logical matchup, since ListByClass has no ORDER BY and so cannot be
-- relied on to always produce Torro1/Torro2 in the same order.

-- Compute a keeper/loser mapping for any pairings that already violate the
-- constraint being added below. On any database where that has never
-- happened (the common case), this finds nothing and every step below is a
-- no-op. The keeper is whichever duplicate has the most Results (i.e. the
-- one people have actually voted on, if any), tie-broken by "Id" for
-- determinism.
CREATE TEMP TABLE pairing_dedup_map AS
WITH ranked AS (
    SELECT
        p."Id",
        LEAST(p."Torro1", p."Torro2") AS lo,
        GREATEST(p."Torro1", p."Torro2") AS hi,
        p."Class",
        ROW_NUMBER() OVER (
            PARTITION BY LEAST(p."Torro1", p."Torro2"), GREATEST(p."Torro1", p."Torro2"), p."Class"
            ORDER BY (SELECT COUNT(*) FROM "Results" WHERE "Pairing" = p."Id") DESC, p."Id" ASC
        ) AS rn
    FROM "Pairings" p
),
keepers AS (
    SELECT lo, hi, "Class", "Id" AS keeper_id FROM ranked WHERE rn = 1
)
SELECT r."Id" AS loser_id, k.keeper_id
FROM ranked r
JOIN keepers k ON k.lo = r.lo AND k.hi = r.hi AND k."Class" = r."Class"
WHERE r.rn > 1;

-- Re-point any Results / AdventVotes referencing a loser duplicate to its
-- keeper before deleting the loser (Results has no ON DELETE, AdventVotes is
-- ON DELETE CASCADE - both would otherwise block or silently lose rows).
UPDATE "Results" SET "Pairing" = m.keeper_id
FROM pairing_dedup_map m
WHERE "Results"."Pairing" = m.loser_id;

UPDATE "AdventVotes" SET "PairingId" = m.keeper_id
FROM pairing_dedup_map m
WHERE "AdventVotes"."PairingId" = m.loser_id;

DELETE FROM "Pairings" WHERE "Id" IN (SELECT loser_id FROM pairing_dedup_map);

DROP TABLE pairing_dedup_map;

-- Enforce it going forward, order-independent.
CREATE UNIQUE INDEX IF NOT EXISTS idx_pairings_unique_matchup
    ON "Pairings" (LEAST("Torro1", "Torro2"), GREATEST("Torro1", "Torro2"), "Class");
