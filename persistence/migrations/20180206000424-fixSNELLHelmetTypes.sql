
-- +migrate Up
update products set document = jsonb_set(document, '{type}', '"helmet"') where document->>'type' != 'helmet';
-- +migrate Down
select 1;