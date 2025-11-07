-- Drop indexes first
DROP INDEX IF EXISTS idx_torrons_active;
DROP INDEX IF EXISTS idx_torrons_new;
DROP INDEX IF EXISTS idx_torrons_gluten_free;
DROP INDEX IF EXISTS idx_torrons_vegan;

-- Drop columns
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "YearAdded";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "Discontinued";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IsNew2025";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IntensityLevel";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IsOrganic";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IsLactoseFree";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IsGlutenFree";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "IsVegan";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "MainIngredients";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "Allergens";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "ProductUrl";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "Price";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "Weight";
ALTER TABLE "Torrons" DROP COLUMN IF EXISTS "Description";
