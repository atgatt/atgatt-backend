-- +migrate Up
create table product_sets (
    id serial primary key,
    uuid uuid not null unique,

    name text null,
    description text null,
    helmet_product_id int null references products(id),
    jacket_product_id int null references products(id),
    pants_product_id int null references products(id),
    boots_product_id int null references products(id),
    gloves_product_id int null references products(id),

    created_at_utc timestamp not null,
    created_by text not null,
    updated_at_utc timestamp null,
    updated_by text null,

    unique(name, helmet_product_id, jacket_product_id, pants_product_id, boots_product_id, gloves_product_id)
);

-- +migrate Down
drop table product_sets;
