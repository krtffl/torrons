-- Remove the Global category.
--
-- "Torrons"."Class" and "Pairings"."Class" reference "Classes" with NO
-- ON DELETE action (RESTRICT), so any child rows scoped to class '5' must be
-- removed before the class row itself, otherwise the final DELETE fails with a
-- foreign-key violation. Remove children before parents:
--   "Results" (references "Pairings"/"Torrons") -> "Pairings" -> "Torrons" -> "Classes".
-- ("Brackets"."ClassId" is ON DELETE CASCADE and is cleaned up automatically.)
-- In normal operation no torró/pairing is tagged with class '5', so these are
-- no-ops and remove no extra data; they only keep the down migration FK-safe.
DELETE FROM "Results"
WHERE "Pairing" IN (SELECT "Id" FROM "Pairings" WHERE "Class" = '5')
   OR "Winner" IN (SELECT "Id" FROM "Torrons" WHERE "Class" = '5');

DELETE FROM "Pairings" WHERE "Class" = '5';

DELETE FROM "Torrons" WHERE "Class" = '5';

DELETE FROM "Classes" WHERE "Id" = '5';
