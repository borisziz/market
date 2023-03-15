-- +goose Up
-- +goose StatementBegin
INSERT INTO STOCKS(warehouse_id, sku, count) VALUES (1,1076963, 10);
INSERT INTO STOCKS(warehouse_id, sku, count) VALUES (2,1076963, 15);
INSERT INTO STOCKS(warehouse_id, sku, count) VALUES (1,1148162, 150); 
INSERT INTO STOCKS(warehouse_id, sku, count) VALUES (1,1625903, 15);


-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
TRUNCATE stocks;
-- +goose StatementEnd
