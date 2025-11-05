alter table "Results" drop constraint if exists fk_results_session;
alter table "Results" drop column if exists "SessionId";
