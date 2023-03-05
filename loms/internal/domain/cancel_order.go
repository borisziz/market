package domain

import (
	"context"

	"github.com/pkg/errors"
)

func (d *domain) CancelOrder(ctx context.Context, orderID int64) error {
	order, err := getOrder(orderID)
	if err != nil {
		return errors.Wrap(err, "get order")
	}
	err = unreserveItems(order.Items)
	if err != nil {
		return errors.Wrap(err, "set items sold")
	}
	err = setOrderStatus(orderID, StatusCancelled)
	if err != nil {
		return errors.Wrap(err, "set order status")
	}
	return nil
}

func unreserveItems(items []OrderItem) error {
	return nil
}
