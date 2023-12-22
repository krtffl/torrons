create table if not exists "Classes"(
    "Id" varchar(36) not null
    constraint pk_classes primary key,
    "Name" varchar(255) not null,
    "Description" text
)
