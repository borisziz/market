package domain

import (
	"context"

	"github.com/pkg/errors"
)

type OrderID int64

var (
	ErrCantReserveItem = errors.New("can not reserve item")
)

func (d *Domain) CreateOrder(ctx context.Context, user int64, items []OrderItem) (OrderID, error) {
	order := Order{Status: StatusNew, User: user, Items: items}
	order.ID = createOrder(order)
	err := reserve(ctx, items)
	if err != nil {
		errSetStatus := setOrderStatus(order.ID, StatusFailed)
		if errSetStatus != nil {
			return 0, errors.Wrap(errSetStatus, "set status after failing reserve items")
		}
		return 0, errors.Wrap(err, "reserve items")
	}
	err = setOrderStatus(order.ID, StatusAwaitingPayment)
	if err != nil {
		return 0, errors.Wrap(err, "set order status")
	}
	return OrderID(order.ID), nil
}

func reserve(ctx context.Context, items []OrderItem) error {
	if false {
		return ErrCantReserveItem
	}
	return nil
}
