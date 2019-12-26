
-- +migrate Up
alter table product_sets drop constraint product_sets_name_helmet_product_id_jacket_product_id_pants_key;
create unique index product_sets_unique_key on product_sets(coalesce(helmet_product_id, -1), coalesce(jacket_product_id, -1), coalesce(pants_product_id, -1), coalesce(boots_product_id, -1), coalesce(gloves_product_id, -1));
-- +migrate Down
alter table product_sets drop constraint product_sets_name_helmet_product_id_jacket_product_id_pants_key;
