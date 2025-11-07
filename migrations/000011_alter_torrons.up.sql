-- Add extended product information to Torrons table
-- This enables richer product display and better filtering options

-- Product details
ALTER TABLE "Torrons"
    ADD COLUMN IF NOT EXISTS "Description" TEXT,
    ADD COLUMN IF NOT EXISTS "Weight" VARCHAR(50),
    ADD COLUMN IF NOT EXISTS "Price" NUMERIC(10,2),
    ADD COLUMN IF NOT EXISTS "ProductUrl" VARCHAR(500);

-- Allergen and ingredient information (stored as text arrays)
ALTER TABLE "Torrons"
    ADD COLUMN IF NOT EXISTS "Allergens" TEXT[],
    ADD COLUMN IF NOT EXISTS "MainIngredients" TEXT[];

-- Dietary and special attributes (boolean flags)
ALTER TABLE "Torrons"
    ADD COLUMN IF NOT EXISTS "IsVegan" BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "IsGlutenFree" BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "IsLactoseFree" BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "IsOrganic" BOOLEAN DEFAULT FALSE;

-- Flavor intensity (1-5 scale)
ALTER TABLE "Torrons"
    ADD COLUMN IF NOT EXISTS "IntensityLevel" INT
        CHECK ("IntensityLevel" IS NULL OR ("IntensityLevel" >= 1 AND "IntensityLevel" <= 5));

-- Product lifecycle tracking
ALTER TABLE "Torrons"
    ADD COLUMN IF NOT EXISTS "IsNew2025" BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "Discontinued" BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS "YearAdded" INT DEFAULT 2025;

-- Create indexes for filtering
CREATE INDEX IF NOT EXISTS idx_torrons_vegan ON "Torrons"("IsVegan") WHERE "IsVegan" = TRUE;
CREATE INDEX IF NOT EXISTS idx_torrons_gluten_free ON "Torrons"("IsGlutenFree") WHERE "IsGlutenFree" = TRUE;
CREATE INDEX IF NOT EXISTS idx_torrons_new ON "Torrons"("IsNew2025") WHERE "IsNew2025" = TRUE;
CREATE INDEX IF NOT EXISTS idx_torrons_active ON "Torrons"("Discontinued") WHERE "Discontinued" = FALSE;
