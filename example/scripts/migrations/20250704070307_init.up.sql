BEGIN;

CREATE TABLE IF NOT EXISTS "users"."user"
(
    id         serial,
    email      varchar(255),
    first_name varchar(255),
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default null,
    deleted_at timestamp with time zone default null
);

COMMIT;