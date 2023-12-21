create table if not exists "Pairings"(
    "Id" varchar(36) not null
    constraint pk_pairings
    primary key,
    "Torro1" varchar(36) not null
    constraint fk_torro_1
    references "Torrons" on delete cascade,
    "Torro2" varchar(36) not null
    constraint fk_torro_2
    references "Torrons" on delete cascade
);
