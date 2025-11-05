create table if not exists "Sessions" (
    "Id" varchar(36) not null primary key,
    "CreatedAt" timestamp not null default current_timestamp,
    "LastSeenAt" timestamp not null default current_timestamp,
    "VoteCount" integer not null default 0,
    "Completed" boolean not null default false
);

create index idx_sessions_created_at on "Sessions"("CreatedAt");
