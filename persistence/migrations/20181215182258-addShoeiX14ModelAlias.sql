
-- +migrate Up
insert into product_model_aliases (manufacturer, model, model_alias) values ('Shoei', 'X Spirit lll', 'X-14');
-- +migrate Down
delete from product_model_aliases where manufacturer = 'Shoei' and model = 'X Spirit lll' and model_alias = 'X-14';
