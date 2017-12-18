
-- +migrate Up
create table products (
    id serial primary key,
    uuid uuid,
    document jsonb,
    created_at date,
    updated_at date
);
-- +migrate Down
drop table products;