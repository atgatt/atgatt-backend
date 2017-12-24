
-- +migrate Up
create table products (
    id serial primary key,
    uuid uuid not null,
    document jsonb not null,
    created_at_utc timestamp not null,
    updated_at_utc timestamp null
);
-- +migrate Down
drop table products;