package domain

import (
	"context"

	"github.com/pkg/errors"
)

var (
	ErrOrderNotFound = errors.New("order not found")
)

func (d *domain) ListOrder(ctx context.Context, orderID int64) (*Order, error) {
	order, err := d.OrdersRepository.GetOrder(ctx, orderID)
	if err != nil {
		return nil, errors.Wrap(err, "get order")
	}
	return order, nil
}
