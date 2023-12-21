create table if not exists "Results"(
    "Id" varchar(36) not null
    constraint pk_results
    primary key,
    "Pairing" varchar(36) not null
    constraint fk_pairing
    references "Pairings",
    "Torro1RatingBefore" numeric not null,
    "Torro2RatingBefore" numeric not null,
    "Winner" varchar(36) not null
    constraint fk_winner
    references "Torrons",
    "Torro1RatingAfter" numeric not null,
    "Torro2RatingAfter" numeric not null
);
