
-- +migrate Up
update product_model_aliases set model = 'Chaser X' where manufacturer = 'Arai' and model = 'Chaser-X';
-- +migrate Down
select 1;