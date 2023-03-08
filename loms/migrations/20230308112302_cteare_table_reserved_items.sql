-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS reserved_items (
    id bigserial PRIMARY KEY,
    order_id integer NOT NULL REFERENCES orders (id),
    warehouse_id integer NOT NULL,
    sku integer NOT NULL,
    count integer NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_reserved_items_sku ON stocks (sku);
CREATE INDEX IF NOT EXISTS idx_reserved_items_warehouse_id ON stocks (warehouse_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_reserved_items_warehouse_id;
DROP INDEX IF EXISTS idx_reserved_items_sku;
DROP TABLE IF EXISTS reserved_items;
-- +goose StatementEnd