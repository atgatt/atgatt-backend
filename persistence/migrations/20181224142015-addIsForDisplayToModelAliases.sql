-- +migrate Up
alter table product_model_aliases add column is_for_display boolean not null default false;
insert into product_model_aliases (manufacturer, model, model_alias) values ('Shoei', 'X Spirit', 'X-11');
update product_model_aliases set is_for_display = true;
update product_model_aliases set is_for_display = false where model_alias in ('X-Fourteen', 'X-Eleven');
create unique index is_for_display_true_once on product_model_aliases (manufacturer, model, is_for_display) where (is_for_display = true);
-- +migrate Down
alter table product_model_aliases drop column is_for_display;
