
-- +migrate Up
insert into product_model_aliases (manufacturer, model, model_alias) values ('Arai','Chaser-X','DT-X');
-- +migrate Down
delete from product_model_aliases where manufacturer = 'Arai' and model_alias = 'DT-X';