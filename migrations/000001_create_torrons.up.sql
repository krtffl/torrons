create table if not exists "Torrons"(
    "Id" varchar(36) not null
    constraint pk_turrons primary key,
    "Name" varchar(255) not null,
    "Rating" numeric not null default 100,
    "Image" varchar(36) not null
);
