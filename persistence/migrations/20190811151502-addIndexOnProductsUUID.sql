
-- +migrate Up
CREATE INDEX IF NOT EXISTS products_uuid_idx ON products (uuid);

-- +migrate Down
DROP INDEX IF EXISTS products_uuid_idx;
