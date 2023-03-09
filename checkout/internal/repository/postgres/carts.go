package repository

import (
	"context"
	"route256/checkout/internal/domain"
	"route256/checkout/internal/repository/schema"
	transactor "route256/libs/postgres_transactor"

	sq "github.com/Masterminds/squirrel"
	"github.com/georgysavva/scany/pgxscan"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
)

var _ domain.CartsRepository = (*CartsRepo)(nil)

type CartsRepo struct {
	transactor.QueryEngineProvider
}

func NewCartsRepo(provider transactor.QueryEngineProvider) *CartsRepo {
	return &CartsRepo{
		QueryEngineProvider: provider,
	}
}

var (
	itemColumns = []string{"sku", "count"}
)

const (
	itemsTable = "cart_items"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

func (r *CartsRepo) GetCartItem(ctx context.Context, user int64, sku uint32) (*domain.CartItem, error) {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := sq.Select(itemColumns...).From(itemsTable).Where(sq.Eq{"user_id": user}).Where(sq.Eq{"sku": sku}).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build orders query")
	}
	var item schema.CartItem
	err = pgxscan.Get(ctx, db, &item, rawQuery, args...)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.CartItem{}, domain.ErrNoSameItemsInCart
		}
		return nil, errors.Wrap(err, "exec orders query")
	}
	return &domain.CartItem{Sku: item.Sku, Count: item.Count}, nil
}

func (r *CartsRepo) GetCart(ctx context.Context, user int64) ([]domain.CartItem, error) {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := sq.Select(itemColumns...).From(itemsTable).Where(sq.Eq{"user_id": user}).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return nil, errors.Wrap(err, "build orders query")
	}
	var items []schema.CartItem
	err = pgxscan.Select(ctx, db, &items, rawQuery, args...)
	if err != nil {
		return nil, errors.Wrap(err, "exec orders query")
	}
	result := make([]domain.CartItem, 0, len(items))
	for _, item := range items {
		result = append(result, domain.CartItem{Sku: item.Sku, Count: item.Count})
	}
	return result, nil
}

func (r *CartsRepo) AddToCart(ctx context.Context, user int64, sku uint32, count uint16) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)

	query := sq.Insert(itemsTable).Columns("user_id", "sku", "count").Values(user, sku, count).Suffix("ON CONFLICT(user_id, sku) DO UPDATE SET count = cart_items.count + ?", count).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build query")
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	return nil
}

func (r *CartsRepo) DeleteFromCart(ctx context.Context, user int64, sku uint32, count uint16, full bool) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	var rawQuery string
	var args []interface{}
	var err error
	if full {
		query := sq.Delete(itemsTable).Where(sq.Eq{"user_id": user}).Where(sq.Eq{"sku": sku}).PlaceholderFormat(sq.Dollar)
		rawQuery, args, err = query.ToSql()
		if err != nil {
			return errors.Wrap(err, "build delete query")
		}
	} else {
		query := sq.Update(itemsTable).Set("count", sq.Expr("count - ?", count)).Where(sq.Eq{"user_id": user}).Where(sq.Eq{"sku": sku}).PlaceholderFormat(sq.Dollar)
		rawQuery, args, err = query.ToSql()
		if err != nil {
			return errors.Wrap(err, "build update query")
		}
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	return nil
}

func (r *CartsRepo) DeleteCart(ctx context.Context, user int64) error {
	db := r.QueryEngineProvider.GetQueryEngine(ctx)
	query := sq.Delete(itemsTable).Where(sq.Eq{"user_id": user}).PlaceholderFormat(sq.Dollar)
	rawQuery, args, err := query.ToSql()
	if err != nil {
		return errors.Wrap(err, "build update query")
	}
	_, err = db.Exec(ctx, rawQuery, args...)
	if err != nil {
		return errors.Wrap(err, "exec query")
	}
	return nil
}
