-- Class "5" was named "Global" at creation (000012), but the leaderboard's
-- category filter row also has its own hardcoded "Global" pill pointing at
-- the synthetic category=global aggregate (all classes combined). Class 5 is
-- a distinct thing: the literal cross-category arena where torrons are
-- entered specifically to compete for "best torró absolute" against other
-- classes. Sharing the name "Global" made the filter row show two
-- identically-labeled buttons pointing at different query params. Renaming
-- class 5 to "Àrbitres" (a name already anticipated in code comments, see
-- internal/domain/persona_stats.go) disambiguates the two without changing
-- either view's behavior.
update "Classes"
set "Name" = 'Àrbitres',
    "Description" = 'L''arena on els millors torrons de cada categoria s''enfronten per determinar el torró absolut'
where "Id" = '5';
