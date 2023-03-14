package repository

import (
	"context"
	transactor "route256/libs/postgres_transactor"
	"route256/loms/internal/domain"
	"route256/loms/internal/repository/schema"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

var _ domain.OrdersRepository = (*OrdersRepo)(nil)

type OrdersRepo struct {
	transactor.QueryEngineProvider
}

func NewItemsRepo(provider transactor.QueryEngineProvider) *OrdersRepo {
	return &OrdersRepo{
		QueryEngineProvider: provider,
	}
}

var (
	ordersColumns = []string{"id", "status", "user_id"}
	itemColumns   = []string{"sku", "count"}
)

const (
	ordersTable        = "orders"
	itemsTable         = "order_items"
	stocksTable        = "stocks"
	reservedItemsTable = "reserved_items"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

func (r *OrdersRepo) CreateOrder(ctx context.Context, order *domain.Order) (int64, error) {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	query := sq.Insert(ordersTable).Columns("status", "user_id").Values(order.Status, order.User).
		Suffix("RETURNING id").PlaceholderFormat(sq.Dollar)

	rawQuery, args, err := query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "build orders query")
	}
	err = pgxscan.Get(ctx, db, &order.ID, rawQuery, args...)
	if err != nil {
		return 0, errors.Wrap(err, "exec orders query")
	}
	query = sq.Insert(itemsTable).Columns("order_id", "sku", "count").PlaceholderFormat(sq.Dollar)
	for _, item := range order.Items {
		query = query.Values(order.ID, item.Sku, item.Count)
	}
	rawQuery, args, err = query.ToSql()
	if err != nil {
		return 0, errors.Wrap(err, "build items query")
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return 0, errors.Wrap(err, "exec items query")
	}
	return order.ID, nil
}

func (r *OrdersRepo) GetOrder(ctx context.Context, orderID int64) (*domain.Order, error) {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := sq.Select(ordersColumns...).
		From(ordersTable).
		Where(sq.Eq{"id": orderID}).PlaceholderFormat(sq.Dollar)

	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build order query")
	}
	var order schema.Order
	err = pgxscan.Get(ctx, db, &order, rawQuery, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrOrderNotFound
		}
		return nil, errors.Wrap(err, "query order")
	}

	query = sq.Select(itemColumns...).
		From(itemsTable).
		Where(sq.Eq{"order_id": orderID}).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err = query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build items query")
	}
	var items []schema.OrderItem
	err = pgxscan.Select(ctx, db, &items, rawQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "query items")
	}
	result := &domain.Order{
		ID:     order.ID,
		User:   order.User,
		Status: order.Status,
		Items:  make([]domain.OrderItem, 0, len(items)),
	}
	for _, item := range items {
		result.Items = append(result.Items, domain.OrderItem{Sku: item.Sku, Count: item.Count})
	}
	return result, nil
}

func (r *OrdersRepo) UpdateOrderStatus(ctx context.Context, id int64, status string, statusNow string) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	query := sq.Update(ordersTable).Set("status", status).
		Where(sq.Eq{"id": id, "status": statusNow}).PlaceholderFormat(sq.Dollar)

	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build query")
	}
	cmd, err := db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	if cmd.RowsAffected() == 0 {
		return ErrOrderNotFound
	}
	return nil
}

func (r *OrdersRepo) ReserveStock(ctx context.Context, orderID int64, item domain.ReservedItem) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := sq.Insert(reservedItemsTable).Columns("order_id", "warehouse_id", "sku", "count").
		Values(orderID, item.WarehouseID, item.Sku, item.Count).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build insert query")
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	return nil
}

func (r *OrdersRepo) UnReserveItems(ctx context.Context, orderID int64) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	query := sq.Delete(reservedItemsTable).Where(sq.Eq{"order_id": orderID}).
		PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build delete query")
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	return nil
}

func (r *OrdersRepo) RemoveSoldedItems(ctx context.Context, orderID int64) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	query := sq.Delete(reservedItemsTable).Where(sq.Eq{"order_id": orderID}).
		Suffix("RETURNING warehouse_id, sku, count").PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build insert query")
	}
	var soldedItems []schema.SoldedItem
	err = pgxscan.Select(ctx, db, &soldedItems, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec insert query")
	}
	b := &pgx.Batch{}
	for _, item := range soldedItems {
		queryDelete := sq.Update(stocksTable).Set("count", sq.Expr("count-?", item.Count)).
			Where(sq.Eq{"warehouse_id": item.WarehouseID, "sku": item.Sku}).PlaceholderFormat(sq.Dollar)
		rawQuery, args, err = queryDelete.ToSql()
		if err != nil {
			return errors.Wrap(err, "build delete query")
		}
		b.Queue(rawQuery, args...)
	}
	br := db.SendBatch(ctx, b)
	defer br.Close()
	for i := 0; i < b.Len(); i++ {
		_, err := br.Exec()
		if err != nil {
			return errors.Wrap(err, "process batch result")
		}
	}
	return nil
}

func (r *OrdersRepo) Stocks(ctx context.Context, sku uint32) ([]domain.Stock, error) {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := `
	SELECT s.warehouse_id, 
		(s.count - (SELECT COALESCE(SUM(r.count), 0) AS count FROM reserved_items r 
			WHERE r.warehouse_id = s.warehouse_id AND sku = s.sku )) AS count 
	FROM stocks s 
		WHERE sku = $1 AND count > 0 ORDER BY count DESC`
	var stocks []schema.Stock
	err := pgxscan.Select(ctx, db, &stocks, query, sku)
	if err != nil {
		return nil, errors.Wrap(err, "exec query stocks")
	}
	var result []domain.Stock
	for _, stock := range stocks {
		result = append(result, domain.Stock{
			WarehouseID: stock.WarehouseID,
			Count:       stock.Count,
		})
	}
	return result, nil
}
