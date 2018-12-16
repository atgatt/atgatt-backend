
-- +migrate Up
update product_model_aliases set is_for_display = false where manufacturer = 'Shoei' and model = 'X Spirit II';
insert into product_model_aliases (manufacturer, model, model_alias, is_for_display) values ('Shoei', 'X Spirit II', 'X-12', true);
-- +migrate Down
delete from product_model_aliases where manufacturer = 'Shoei' and model = 'X Spirit II' and model_alias = 'X-12';
