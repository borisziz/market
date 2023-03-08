-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS orders 
(
    id bigserial PRIMARY KEY,
    status text NOT NULL,
    user_id integer NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS orders;
-- +goose StatementEnd
