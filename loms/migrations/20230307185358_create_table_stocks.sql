-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS stocks (
    id bigserial,
    warehouse_id integer NOT NULL,
    sku integer NOT NULL,
    count integer NOT NULL,
    PRIMARY KEY(warehouse_id, sku)
);
CREATE INDEX IF NOT EXISTS idx_stocks_sku ON stocks (sku);
CREATE INDEX IF NOT EXISTS idx_stocks_warehouse_id ON stocks (warehouse_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_stocks_warehouse_id;
DROP INDEX IF EXISTS idx_stocks_sku;
DROP TABLE IF EXISTS stocks;
-- +goose StatementEnd
