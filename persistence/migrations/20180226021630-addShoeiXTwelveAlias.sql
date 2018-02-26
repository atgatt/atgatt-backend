
-- +migrate Up
insert into product_model_aliases (manufacturer, model, model_alias) values ('Shoei','X Spirit II','X-Twelve');
-- +migrate Down
delete from product_model_aliases where manufacturer = 'Shoei' and model_alias = 'X-Twelve';