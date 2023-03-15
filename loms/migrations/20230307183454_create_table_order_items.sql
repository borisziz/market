-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS order_items (
    id bigserial PRIMARY KEY,
    order_id bigint NOT NULL REFERENCES orders (id),
    sku integer NOT NULL,
    count int4 NOT NULL
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS reserved_items;
-- +goose StatementEnd
