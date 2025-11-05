alter table "Results"
add column "SessionId" varchar(36);

alter table "Results"
add constraint fk_results_session
foreign key ("SessionId") references "Sessions"("Id") on delete set null;

create index idx_results_session_id on "Results"("SessionId");
