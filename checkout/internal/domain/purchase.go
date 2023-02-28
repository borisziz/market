package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (m *domain) Purchase(ctx context.Context, user int64) (int64, error) {
	orderID, err := m.lOMSCaller.CreateOrder(ctx, user, []CartItem{{Sku: 1076963, Count: 1}, {Sku: 1148162, Count: 3}})
	if err != nil {
		return orderID, errors.WithMessage(err, "creating order")
	}
	return orderID, nil
}
