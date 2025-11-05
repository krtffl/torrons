-- Add Global category for cross-category comparisons
-- This allows comparing torrons from different categories to create
-- an overall ranking across all types

INSERT INTO "Classes"
    ("Id", "Name", "Description")
VALUES
    ('5', 'Global', 'Comparació entre totes les categories per determinar el millor torró absolut');
