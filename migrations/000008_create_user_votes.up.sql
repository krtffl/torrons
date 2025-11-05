create table if not exists "UserVotes" (
    "Id" varchar(36) not null primary key,
    "SessionId" varchar(36) not null,
    "PairingId" varchar(36) not null,
    "WinnerId" varchar(36) not null,
    "VotedAt" timestamp not null default current_timestamp,
    foreign key ("SessionId") references "Sessions"("Id") on delete cascade,
    foreign key ("PairingId") references "Pairings"("Id") on delete cascade,
    foreign key ("WinnerId") references "Torrons"("Id") on delete cascade
);

create index idx_user_votes_session_id on "UserVotes"("SessionId");
create index idx_user_votes_voted_at on "UserVotes"("VotedAt");
